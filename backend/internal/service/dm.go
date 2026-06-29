package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
)

// GetOrCreateDM returns the channel ID of an existing DM between two users,
// or creates one if none exists.
func GetOrCreateDM(ctx context.Context, queries *db.Queries, rawDB *sql.DB, userA, userB uuid.UUID) (uuid.UUID, error) {
	existing, err := queries.FindExistingDM(ctx, db.FindExistingDMParams{
		UserID:   userA,
		UserID_2: userB,
	})
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return uuid.Nil, err
	}

	// Create in a transaction so we never get a half-created DM.
	tx, err := rawDB.BeginTx(ctx, nil)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback()

	qtx := queries.WithTx(tx)

	channel, err := qtx.CreateChannel(ctx, db.CreateChannelParams{
		Name:        "",
		Description: sql.NullString{},
		CreatedBy:   userA,
	})
	if err != nil {
		return uuid.Nil, err
	}

	// Set kind to 'dm' — CreateChannel uses the default 'public', so we
	// update it immediately within the same transaction.
	if _, err := tx.ExecContext(ctx,
		"UPDATE channels SET kind = 'dm' WHERE id = $1", channel.ID,
	); err != nil {
		return uuid.Nil, err
	}

	for _, uid := range []uuid.UUID{userA, userB} {
		if err := qtx.AddChannelMember(ctx, db.AddChannelMemberParams{
			ChannelID: channel.ID,
			UserID:    uid,
		}); err != nil {
			return uuid.Nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return uuid.Nil, err
	}
	return channel.ID, nil
}
