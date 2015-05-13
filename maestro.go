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
}

func (m *Maestro) portalReconnect() error {
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

func (m *Maestro) portalReceive() error {
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

func (m *Maestro) portalEmitExistance() error {
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
