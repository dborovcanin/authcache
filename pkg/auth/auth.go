package auth

import (
	"context"

	"github.com/mainflux/mainflux"
)

type auth struct {
	cache *cache
	tc    mainflux.ThingsServiceClient
}

func (a *auth) Identify(password string) (string, error) {
	t := &mainflux.Token{
		Value: string(password),
	}

	thid, err := a.tc.Identify(context.TODO(), t)
	if err != nil {
		return "", err
	}

	return thid.GetValue(), nil
}

func (a *auth) Authorize(thingID, channelID string) error {
	if a.cache.Validate(thingID, channelID) {
		return true
	}
	ar := &mainflux.AccessByIDReq{
		ThingID: thingID,
		ChanID:  chanID,
	}

	_, err := a.tc.CanAccessByID(context.TODO(), ar)
	if err != nil {
		a.cache.Add(thingID, channelID)
	}
	return err
}
