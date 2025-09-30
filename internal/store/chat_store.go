package store

import (
	"context"
	"myGinServer/models/chat_msg"
)

func (s *DbStore) ContactList(ctx context.Context, userId string) ([]*chat_msg.Contact, error) {
	sqlText := `select id, owner_id, target_id, type, desc from contact where owner_id=?`
	contacts := make([]*chat_msg.Contact, 0)
	err := s.db.SelectContext(ctx, &contacts, sqlText, userId)
	return contacts, err
}

func (s *DbStore) ContactSave(ctx context.Context, contact *chat_msg.Contact) error {
	sqlText := `insert into contact(owner_id, target_id, type, desc) values (?,?,?,?)`
	_, err := s.db.ExecContext(ctx, sqlText, contact.OwnerId, contact.TargetId, contact.Type, contact.Desc)
	return err
}
