package service

import (
	"context"
	"time"

	"github.com/bots-house/share-file-bot/core"
	"github.com/friendsofgo/errors"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
)

type Auth struct {
	UserStore core.UserStore
}

type UserInfo struct {
	ID           int
	FirstName    string
	LastName     string
	Username     string
	LanguageCode string
}

func (srv *Auth) createUser(ctx context.Context, info *UserInfo) (*core.User, error) {
	user := core.NewUser(
		core.UserID(info.ID),
		info.FirstName,
		info.LastName,
		info.Username,
		info.LanguageCode,
	)

	log.Ctx(ctx).Info().Msg("create new user")
	if err := srv.UserStore.Add(ctx, user); err != nil {
		return nil, errors.Wrap(err, "add user to store")
	}

	return user, nil
}

func (srv *Auth) updateUserIfNeed(ctx context.Context, user *core.User, info *UserInfo) (*core.User, error) {
	var update bool

	if user.FirstName != info.FirstName {
		user.FirstName = info.FirstName
		update = true
	}

	if user.LastName.String != info.LastName {
		user.LastName = null.NewString(info.LastName, info.LastName != "")
		update = true
	}

	if user.Username.String != info.Username {
		user.Username = null.NewString(info.Username, info.Username != "")
		update = true
	}

	if !update {
		return user, nil
	}

	user.UpdatedAt = null.TimeFrom(time.Now())

	log.Ctx(ctx).Info().Msg("update user info")
	if err := srv.UserStore.Update(ctx, user); err != nil {
		return nil, errors.Wrap(err, "update user in store")
	}

	return user, nil
}

func (srv *Auth) Auth(ctx context.Context, info *UserInfo) (*core.User, error) {
	id := core.UserID(info.ID)

	user, err := srv.UserStore.Find(ctx, id)
	if err == core.ErrUserNotFound {
		user, err := srv.createUser(ctx, info)
		if err != nil {
			return nil, errors.Wrap(err, "create user")
		}
		return user, nil
	} else if err != nil {
		return nil, errors.Wrap(err, "find user")
	}

	user, err = srv.updateUserIfNeed(ctx, user, info)
	if err != nil {
		return nil, errors.Wrap(err, "fail to update user")
	}

	return user, nil
}

func (srv *Auth) SettingsToggleLongIDs(ctx context.Context, user *core.User) (bool, error) {
	updated := user.Settings.Patch(func(settings *core.UserSettings) {
		settings.LongIDs = !settings.LongIDs
	})

	if updated {
		if err := srv.UserStore.Update(ctx, user); err != nil {
			return false, errors.Wrap(err, "update user")
		}

	}

	return user.Settings.LongIDs, nil
}
