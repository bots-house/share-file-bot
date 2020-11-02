package service

import (
	"context"

	"github.com/bots-house/share-file-bot/core"
	"github.com/friendsofgo/errors"
	"golang.org/x/sync/errgroup"
)

type Admin struct {
	User     core.UserStore
	File     core.FileStore
	Download core.DownloadStore
	Chat     core.ChatStore
}

type AdminSummaryStats struct {
	Users     int
	Files     int
	Downloads int
	Chats     int

	UsersByRefs core.UserRefStats
}

var (
	ErrUserIsNotAdmin = errors.New("user is not admin")
	ErrArgsAreEmpty   = errors.New("command args are empty")
)

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
		docs, err := srv.File.Query().Count(ctx)
		if err != nil {
			return errors.Wrap(err, "count files")
		}

		stats.Files = docs

		return nil
	})

	wg.Go(func() error {
		dwns, err := srv.Download.Query().Count(ctx)
		if err != nil {
			return errors.Wrap(err, "count downloads")
		}

		stats.Downloads = dwns

		return nil
	})

	wg.Go(func() error {
		chats, err := srv.Chat.Query().Count(ctx)
		if err != nil {
			return errors.Wrap(err, "count chats")
		}

		stats.Chats = chats

		return nil
	})

	wg.Go(func() error {
		refs, err := srv.User.RefStats(ctx)
		if err != nil {
			return errors.Wrap(err, "count user refs")
		}

		stats.UsersByRefs = refs

		return nil
	})

	if err := wg.Wait(); err != nil {
		return nil, err
	}

	return stats, nil
}

func (srv *Admin) isHasPermissions(_ context.Context, user *core.User) error {
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

func (srv *Admin) SummaryRefStats(ctx context.Context, user *core.User, ref string) (*core.SummaryRefStats, error) {
	if err := srv.isHasPermissions(ctx, user); err != nil {
		return nil, err
	}

	summary, err := srv.User.SummaryRefStats(ctx, ref)
	if err != nil {
		return nil, errors.Wrap(err, "get ref stats")
	}

	return summary, nil
}
