package message

import (
	"github.com/grumblechat/server/internal/pagination"

	"github.com/segmentio/ksuid"
	bolt "go.etcd.io/bbolt"
)

func GetAll(db *bolt.DB, channelID ksuid.KSUID, pgn pagination.Pagination) ([]*Message, error) {
	var messages []*Message

	err := db.View(func(tx *bolt.Tx) error {
		dbb := channelBucket(tx, channelID)
		csr := dbb.Cursor()
		var ctr uint16 = 1

		// iterate over all messages, decode, and add to result
		for k, v := pgn.InitCursor(csr); ctr <= pgn.Count && k != nil; k, v = pgn.MoveCursor(csr) {
			decoded, err := Decode(v)
			if err != nil {
				return err
			}
			messages = append(messages, decoded)
			ctr++
		}

		endKey, err := pgn.EndKey(csr)
		if err != nil {
			return err
		}

		pgn.NextCursor = endKey
		return nil
	})

	return messages, err
}

func Find(db *bolt.DB, channelID ksuid.KSUID, id ksuid.KSUID) (*Message, error) {
	var msg *Message

	err := db.View(func(tx *bolt.Tx) error {
		dbb := channelBucket(tx, channelID)

		// get by ID
		res := dbb.Get(id.Bytes())
		if res == nil {
			msg = nil
			return nil
		}

		// decode channel
		decoded, err := Decode(res)
		if err != nil {
			return err
		}

		msg = decoded
		return nil
	})

	return msg, err
}
