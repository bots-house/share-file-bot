package service

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/friendsofgo/errors"
	"golang.org/x/sync/errgroup"
)

type Admin struct {
	User     core.UserStore
	Document core.DocumentStore
	Download core.DownloadStore
}

type AdminSummaryStats struct {
	Users     int
	Documents int
	Downloads int
}

var ErrUserIsNotAdmin = errors.New("user is not admin")

func (srv *Admin) getStats(ctx context.Context) (*AdminSummaryStats, error) {
	stats := &AdminSummaryStats{}

	wg, ctx := errgroup.WithContext(ctx)

	wg.Go(func() error {
		users, err := srv.User.Query().Count(ctx)
		if err != nil {
			return errors.Wrap(err, "count users")
		}

		stats.Users = users

		return nil
	})

	wg.Go(func() error {
		docs, err := srv.Document.Query().Count(ctx)
		if err != nil {
			return errors.Wrap(err, "count docs")
		}

		stats.Documents = docs

		return nil
	})

	wg.Go(func() error {
		dwns, err := srv.Download.Query().Count(ctx)
		if err != nil {
			return errors.Wrap(err, "count docs")
		}

		stats.Downloads = dwns

		return nil
	})

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return stats, nil
}

func (srv *Admin) isHasPermissions(ctx context.Context, user *core.User) error {
	if !user.IsAdmin {
		return ErrUserIsNotAdmin
	}

	return nil
}

func (srv *Admin) SummaryStats(ctx context.Context, user *core.User) (*AdminSummaryStats, error) {
	if err := srv.isHasPermissions(ctx, user); err != nil {
		return nil, err
	}

	stats, err := srv.getStats(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get stats")
	}

	return stats, nil
}
