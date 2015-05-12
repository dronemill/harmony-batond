package main

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/dronemill/harmony-client-go"
	"github.com/fsouza/go-dockerclient"
)

// Batond is the main app contianer
type Batond struct {
	// hamrony client
	Harmony *harmonyclient.Client
}

// maestroConnect will get a connected maestro client
func (b *Batond) harmonyConnect() error {
	hconf := harmonyclient.Config{
		APIHost:      config.Harmony.API,
		APIVersion:   "v1",
		APIVerifySSL: config.Harmony.VerifySSL,
	}

	log.WithField("harmonyAPI", config.Harmony.API).Info("Attempting connection to HarmonyAPI")

	var err error
	b.Harmony, err = harmonyclient.NewHarmonyClient(hconf)

	if err != nil {
		// TODO: maybe like dont bomb out here.. @pmccarren
		log.Fatalf("Failed connecting to the HarmonyAPI: %s", err.Error())
	}

	return nil
}

// getMachine will get the Harmony Machine
func (b *Batond) getMachine() *harmonyclient.Machine {
	// check if we already have a MachineID
	machine := b.getMachineByName()
	if machine != nil {
		return machine
	}

	return b.createMachine()
}

// getMachineByName will get a Harmony Machine resource by name
func (b *Batond) getMachineByName() *harmonyclient.Machine {
	machine, err := b.Harmony.MachineByName(config.Machine.Name)
	if err != nil {
		// TODO: need to get some better error handling in here.. @pmccarren
		log.WithField("name", config.Machine.Name).WithField("error", err.Error()).Fatal("Failed fetching machine by name")
	}

	return machine
}

// createMachine will create a Harmony Machine resource
func (b *Batond) createMachine() *harmonyclient.Machine {

	machine := &harmonyclient.Machine{
		Name:     config.Machine.Name,
		Hostname: config.Machine.Hostname,
	}

	machineResource, err := b.Harmony.MachinesAdd(machine)

	if err != nil {
		log.WithField("error", err.Error()).Fatal("Failed creating machine")
	}

	return machineResource
}

/**
 * Fetch container from Harmony
 *
 * Check if the container is already created
 * 		IF IS NOT; check if image is pulled
 * 			IF NO:: pull
 * 		create it and store the cID w/ Harmony
 *
 * Check if container is enabled
 * 		IF IS: ensure started
 * 		IF IS: NOT ensure stopped
 */
// checkContainerState will ensure a container is in the propper state
func (b *Batond) checkContainerState(cID string) {
	// first we need to get the container resource
	container, err := b.Harmony.Container(cID)
	if err != nil {
		// TODO: need to get some better error handling in here.. @pmccarren
		log.WithField("containerID", cID).WithField("error", err.Error()).Fatal("Failed fetching container")
	}
	log.WithField("containerID", container.ID).Info("Successfuly retreived Harmony container resource")

	// check if the container is already created
	exists, err := b.containerExists(container)
	if err != nil {
		log.WithField("containerID", cID).WithField("error", err.Error()).Fatal("Failed checking if container exists")
	}

	// FIXME: need to handle the case wheere the container.CID is set, but the container does not actually exist

	if !exists {
		log.WithField("containerID", container.ID).Info("Container does not exist. Creating.")

		// ensure image exists
		if err := b.ensureImageExists(container.Image); err != nil {
			log.WithField("containerID", cID).WithField("image", container.Image).WithField("error", err.Error()).Fatal("Failed ensuring image exists")
		}

		dkrContainer, err := b.createContainer(container)
		if err != nil {
			log.WithField("containerID", cID).WithField("error", err.Error()).Fatal("Failed creating container")
		}

		log.WithField("containerID", cID).WithField("cID", dkrContainer.ID).Info("Created container")

		if err := b.Harmony.ContainersCIDUpdate(cID, dkrContainer.ID); err != nil {
			log.WithField("containerID", cID).WithField("cID", dkrContainer.ID).WithField("error", err.Error()).Fatal("Failed updating container")
		}

		// manually set it here, so we dont have to refetch the container resource
		container.CID = dkrContainer.ID

		log.WithField("containerID", cID).WithField("cID", dkrContainer.ID).Info("Updated container cID")
	} else {
		log.WithField("containerID", container.ID).WithField("cID", container.CID).Debug("Container already exists")
	}

	// get the running state of the container
	running, err := b.containerRunning(container)
	if err != nil {
		log.WithField("containerID", cID).WithField("cID", container.CID).WithField("error", err.Error()).Fatal("Failed checking if container is running")
	}

	log.WithField("containerID", cID).WithField("cID", container.CID).WithField("running", running).Debug("Checked container running state")

	// if container is stopped, and should be running, start it
	if !running && container.Enabled {
		if err := b.startContainer(container); err != nil {
			log.WithField("containerID", cID).WithField("cID", container.CID).WithField("error", err.Error()).Fatal("Failed starting container")
		}

		log.WithField("containerID", cID).WithField("cID", container.CID).WithField("running", true).Info("Started container")
	}

	// if container is running, and should be stopped, stop it
	if running && !container.Enabled {
		if err := b.stopContainer(container); err != nil {
			log.WithField("containerID", cID).WithField("cID", container.CID).WithField("error", err.Error()).Fatal("Failed stopping container")
		}

		log.WithField("containerID", cID).WithField("cID", container.CID).WithField("running", false).Info("Stopped container")
	}
}

// ensureImageExists will make sure that we have an image pulled
func (b *Batond) ensureImageExists(name string) error {
	// get the image id
	imageID, err := b.dkrImageNameToID(name)

	if err != nil {
		log.WithField("image", name).WithField("error", err.Error()).Error("Failed checking if image is already pulled")
		return err
	}

	if imageID != "" {
		log.WithField("image", name).
			Debug("Image is already pulled")
		return nil
	}

	log.WithField("image", name).Info("Pulling image")

	var registry, image, tag string

	// We need to split up the image name
	parts := strings.Split(name, ":")
	image = parts[0]

	// if we have a tag, then set it
	if len(parts) > 1 {
		tag = parts[1]
	}

	// check and see if we have two slashes, because if we do,
	// then preceeding the first one is the registry
	if count := strings.Count(image, "/"); count == 2 {
		imageParts := strings.Split(image, "/")
		registry = imageParts[0]
	}

	opts := docker.PullImageOptions{
		Registry:   registry,
		Repository: image,
		Tag:        tag,
	}

	// FIXME
	// auth := docker.AuthConfiguration{
	// 	Username:      conf.DockerRegistry.Username,
	// 	Password:      conf.DockerRegistry.Password,
	// 	Email:         conf.DockerRegistry.Email,
	// 	ServerAddress: conf.DockerRegistry.Url,
	// }
	auth := docker.AuthConfiguration{}

	// Pull the image
	err = dkr.PullImage(opts, auth)

	if err != nil {
		log.WithField("image", name).WithField("error", err.Error()).Error("Failed pulling image")
		return err
	}

	log.WithField("image", name).
		Debug("Pulled image")

	return nil
}

// dkrImageNameToID will convery an image name to the docker imageID
func (b *Batond) dkrImageNameToID(name string) (string, error) {
	images, err := dkr.ListImages(docker.ListImagesOptions{})

	if err != nil {
		log.Error(err.Error())
		return "", err
	}

	// if there isn't a defined tag, then we are referring to 'latest'
	if !strings.Contains(name, ":") {
		name = name + ":latest"
	}

	for _, i := range images {
		for _, tag := range i.RepoTags {
			if tag == name {
				return i.ID, nil
			}
		}
	}

	return "", nil
}

// containerExists will check if the container is already created
func (b *Batond) containerExists(container *harmonyclient.Container) (bool, error) {
	if container.CID == "" {
		return false, nil
	}

	// see if the container is already created
	dkrContainer, err := dkr.InspectContainer(container.CID)
	if err != nil {
		if _, ok := err.(*docker.NoSuchContainer); ok {
			return false, nil
		}

		log.WithField("containerID", container.ID).WithField("error", err.Error()).Fatal("Failed inspecting container")
		return false, err
	}

	// if we received a container, then return so!
	if dkrContainer != nil {
		return true, nil
	}

	return false, nil
}

// containerRunning will check if the container is running
func (b *Batond) containerRunning(container *harmonyclient.Container) (bool, error) {
	// see if the container is already created
	dkrContainer, err := dkr.InspectContainer(container.CID)
	if err != nil {
		log.WithField("containerID", container.ID).WithField("error", err.Error()).Fatal("Failed inspecting container")
		return false, err
	}

	return dkrContainer.State.Running, nil
}

// createContainer will create a new docker container
func (b *Batond) createContainer(container *harmonyclient.Container) (*docker.Container, error) {
	log.WithField("containerID", container.ID).Info("Creating container")

	// setup some vars
	var entryPoint, cmd, env []string

	// set EntryPoint if exists
	if container.EntryPoint != "" {
		entryPoint = make([]string, 1)
		entryPoint[0] = container.EntryPoint
	}

	// set Cmd if exists
	if container.Cmd != "" {
		cmd = strings.Split(container.Cmd, " ")
	}

	// compile the env
	env = make([]string, len(container.ContainerEnvsIDs))
	for i, eID := range container.ContainerEnvsIDs {
		e, err := b.Harmony.ContainerEnv(eID)
		if err != nil {
			log.WithField("ContainerEnvID", eID).WithField("error", err.Error()).Fatalf("Failed retreiving env")
		}

		// TODO excape this better
		env[i] = fmt.Sprintf("%s=\"%s\"", e.Name, e.Value)
	}

	hostConfig, err := b.containerHostConfig(container)

	if err != nil {
		log.WithField("ContainerID", container.ID).WithField("error", err.Error()).Fatalf("Failed building hostConfig")
	}

	dkrConfig := docker.Config{
		Hostname:   container.Hostname,
		Env:        env,
		Entrypoint: entryPoint,
		Cmd:        cmd,
		DNS:        hostConfig.DNS,
		Image:      container.Image,
		Tty:        container.Tty,
		OpenStdin:  container.Interactive,
		StdinOnce:  container.Interactive,
	}

	opts := docker.CreateContainerOptions{
		Name:       container.Name,
		Config:     &dkrConfig,
		HostConfig: &hostConfig,
	}

	c, err := dkr.CreateContainer(opts)

	if err != nil {
		log.WithField("containerID", container.ID).WithField("error", err.Error()).Error("Failed creating docker container")

		return nil, err
	}

	log.WithField("containerID", container.ID).WithField("cID", c.ID).Info("Created docker container")

	return c, nil
}

// containerHostConfig will generate a container's config by fetching the relevant
// resources from Harmony, and composing the env and dns slices
func (b *Batond) containerHostConfig(container *harmonyclient.Container) (docker.HostConfig, error) {
	// compile the dns
	dns := make([]string, len(container.ContainerDnsIDs))
	for i, dID := range container.ContainerDnsIDs {
		d, err := b.Harmony.ContainerDns(dID)
		if err != nil {
			log.WithField("ContainerDnsID", dID).WithField("error", err.Error()).Fatal("Failed retreiving container dns")
		}

		dns[i] = d.Nameserver
	}

	// compile the binds
	binds := make([]string, len(container.ContainerVolumesIDs))
	for i, vID := range container.ContainerVolumesIDs {
		v, err := b.Harmony.ContainerVolume(vID)
		if err != nil {
			log.WithField("ContainerVolumeID", vID).WithField("error", err.Error()).Fatal("Failed retreiving container volume")
		}

		binds[i] = fmt.Sprintf("%s:%s", v.PathHost, v.PathContainer)
	}

	log.WithField("containerID", container.ID).Debug("Successfully built containerHostConfig")

	return docker.HostConfig{
		Binds: binds,
		DNS:   dns,
	}, nil
}

// startContainer will start a docker container
func (b *Batond) startContainer(container *harmonyclient.Container) error {
	hostConfig, err := b.containerHostConfig(container)
	if err != nil {
		log.WithField("ContainerID", container.ID).WithField("error", err.Error()).Fatalf("Failed building hostConfig")
	}

	err = dkr.StartContainer(container.CID, &hostConfig)
	if err != nil {
		log.WithField("containerID", container.ID).WithField("error", err.Error()).Error("Failed starting container")
	}

	return err
}

// stopContainer will start a docker container
func (b *Batond) stopContainer(container *harmonyclient.Container) error {
	// FIXME
	var timeout uint
	timeout = 5

	err := dkr.StopContainer(container.CID, timeout)
	if err != nil {
		log.WithField("containerID", container.ID).WithField("error", err.Error()).Error("Failed stopping container")
	}

	return err
}
