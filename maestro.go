package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/dronemill/eventsocket-client-go"
	"github.com/dronemill/harmony-client-go"
)

// Maestro is the main app contianer
type Maestro struct {
	// Harmony client
	Harmony *harmonyclient.Client

	// Portal is the Portal client instance
	Portal *eventsocketclient.Client

	// batondChanContainerUpdate is the channel on which to
	// send updated containerIDs
	batondChanContainerUpdate chan string
}

// PortalReconnect reconnects to the portal after closing connections
func (m *Maestro) PortalReconnect() error {
	err := m.Portal.Reconnect()

	if err != nil {
		log.WithField("maestro", config.Maestro.Host).
			WithField("portalPort", config.Maestro.PortalPort).
			WithField("error", err.Error()).
			Fatal("Failed reconnecting to the Maestro Portal")
		return err
	}

	return nil
}

// Suscribe to container events
func (m *Maestro) Suscribe() {
	go m.suscribeContainerUpdate()
}

// suscribeContainerUpdate will subscribe to, and handle events from batond-container-updated
func (m *Maestro) suscribeContainerUpdate() {
	event := fmt.Sprintf("harmony.machine-%s.batond-container-updated", machine.ID)
	containerUpdateChan, err := m.Portal.Suscribe(event)

	if err != nil {
		log.WithField("error", err.Error()).
			WithField("event", "batond_boot").
			Fatal("Failed suscribing to event")
	}
	log.WithField("event", "batond_boot").
		Info("Suscribed to event")

	for {
		select {
		case msg := <-containerUpdateChan:
			m.handleContainerUpdate(msg)
		}
	}
}

// PortalReceive will handle incomming events from the portal
func (m *Maestro) PortalReceive() error {
	log.Info("Starting portal receive routine")

	go m.Portal.Recv()

	for {
		select {
		// case req := <-m.Portal.RecvRequest:
		// 	if req.Err != nil {
		// 		panic(req.Err)
		// 	}
		// 	handleRequest(req)
		case err := <-m.Portal.RecvError:
			// FIXME make logging better
			log.Error(fmt.Sprintf("Failed receiving from the websocket: %s", err.Error()))
			return err
		}
	}
}

// PortalEmitExistance tells the portal about self
func (m *Maestro) PortalEmitExistance() error {
	log.WithField("maestro", config.Maestro.Host).WithField("portalPort", config.Maestro.PortalPort).Debug("Emitting existance")

	p := eventsocketclient.NewPayload()
	p["ClientID"] = m.Portal.Id
	p["MachineID"] = machine.ID

	err := m.Portal.Emit("batond_boot", &p)
	if err != nil {
		log.WithField("maestro", config.Maestro.Host).
			WithField("portalPort", config.Maestro.PortalPort).
			WithField("error", err.Error()).
			Fatal("Failed emitting existance")

		return err
	}

	log.WithField("maestro", config.Maestro.Host).WithField("portalPort", config.Maestro.PortalPort).Debug("Successfully emitted existance")
	return nil
}

// handleContainerUpdate will update a changed container's state
func (m *Maestro) handleContainerUpdate(r *eventsocketclient.Received) {
	containerID := (*r.Message.Payload)["ContainerID"].(string)

	log.WithField("maestro", config.Maestro.Host).WithField("containerID", containerID).Info("Handling batond-container-update")

	m.batondChanContainerUpdate <- containerID
}
