package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository interface {
	Create(ctx context.Context, usernames []string) (int64, error)
	Delete(ctx context.Context, id int64) error
	SendMessage(ctx context.Context, msg MessageCreateDTO) error
}

type repo struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) ChatRepository {
	return &repo{
		db: db,
	}
}

func (r *repo) Create(ctx context.Context, usernames []string) (int64, error) {

	// Итак, логика такая, сначала создаем чат, то есть делаем запись в таблицу chats, после этого возвращается id чата, который мы будем указывать в таблице chat_members
	// builder := squirrel.Insert("chats").
	// 	PlaceholderFormat(squirrel.Dollar).
	// 	Columns("created_at").
	// 	Values(time.Now()).
	// 	Suffix("RETURNING id")

	// query, args, err := builder.ToSql()
	// if err != nil {
	// 	return 0, err
	// }
	// // Создали запись в таблице chats и получили ее id
	// var chat_id int64
	// err = r.db.QueryRow(ctx, query, args...).Scan(&chat_id)
	// if err != nil {
	// 	return 0, err
	// }

	var chatId int64
	// Здесь используем простой запрос, потому что у нас для chats по умолчанию дефолтные значения, поэтому не нужно дополнительно что то вставлять
	sql := "INSERT INTO chats DEFAULT VALUES RETURNING id"
	err := r.db.QueryRow(ctx, sql).Scan(&chatId)
	if err != nil {
		return 0, err
	}

	// Теперь нужно создать записи в таблице chat_members, в которых будем указывать id только что созданного чата
	for _, username := range usernames {
		inserMembers := squirrel.Insert("chat_members").
			PlaceholderFormat(squirrel.Dollar).
			Columns("chat_id", "username", "joined_at").
			Values(chatId, username, time.Now())

		query, args, err := inserMembers.ToSql()
		if err != nil {
			return 0, err
		}

		_, err = r.db.Exec(ctx, query, args...)
		if err != nil {
			return 0, err
		}
	}

	return chatId, nil

}

func (r *repo) Delete(ctx context.Context, chat_id int64) error {
	builder := squirrel.Delete("chats").
		PlaceholderFormat(squirrel.Dollar).
		Where(squirrel.Eq{"id": chat_id})

	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}

func (r *repo) SendMessage(ctx context.Context, msg MessageCreateDTO) error {
	// Добавим валидацию, что пользователь есть в чате, делаю запрос без использования squirrel
	var isMember bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM chat_members WHERE chat_id = $1 AND username = $2)", msg.Chat_id, msg.From_username).Scan(&isMember)
	if err != nil {
		return err
	}

	if !isMember {
		return errors.New("user is not a member of the chat")
	}

	// Выполняем основной запрос, добавляем в таблицу messages
	builder := squirrel.Insert("messages").
		PlaceholderFormat(squirrel.Dollar).
		Columns("chat_Id", "from_username", "text").
		Values(msg.Chat_id, msg.From_username, msg.Text)
	query, args, err := builder.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, query, args...)
	if err != nil {
		return err
	}

	return nil
}
