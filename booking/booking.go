// package
package booking

import (
	"context"
	"time"

	"encore.app/booking/db"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
)

var (
	bookingDB = sqldb.NewDatabase("booking", sqldb.DatabaseConfig{
		Migrations: "./db/migrations",
	})

	pgxdb = sqldb.Driver[*pgxpool.Pool](bookingDB)
	query = db.New(pgxdb)
)

type Booking struct {
	ID    int64     `json:"id"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
	Email string    `encore:"sensitive"`
}

type BookParams struct {
	Start time.Time `json:"start"`
	Email string    `encore:"sensitive"`
}

//encore:api public method=POST path=/booking
func Book(ctx context.Context, p *BookParams) error {
	eb := errs.B()

	now := time.Now()
	if p.Start.Before(now) {
		return eb.Code(errs.InvalidArgument).Msg("start time must be in the future").Err()
	}

	tx, err := pgxdb.Begin(ctx)
	if err != nil {
		return eb.Cause(err).Code(errs.Unavailable).Msg("failed to start transaction").Err()
	}
	defer tx.Rollback(context.Background()) // committed explicitly on success

	_, err = query.InsertBooking(ctx, db.InsertBookingParams{
		StartTime: pgtype.Timestamp{Time: p.Start, Valid: true},
		EndTime:   pgtype.Timestamp{Time: p.Start.Add(DefaultBookingDuration), Valid: true},
		Email:     p.Email,
	})
	if err != nil {
		return eb.Cause(err).Code(errs.Unavailable).Msg("failed to insert booking").Err()
	}

	if err := tx.Commit(ctx); err != nil {
		return eb.Cause(err).Code(errs.Unavailable).Msg("failed to commit transaction").Err()
	}
	return nil
}
