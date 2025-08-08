package repository

import (
	"chat-server/internal/repository/model"
	"context"
	"errors"
	"log"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ChatRepository interface {
	Create(ctx context.Context, usernames []string) (int64, error)
	Delete(ctx context.Context, id int64) error
	SendMessage(ctx context.Context, msg model.Message) error
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
	// Проверяем что req.Usernames не пусто
	if len(usernames) == 0 {
		return 0, errors.New("usernames cannot be empty")
	}
	// Пока будем просто будем создавать чат с переданными usernames, дальше нужно будет изменить
	// TODO: первым делом нужно будет проверить что пользователи существуют, с помощью запроса к сервису AUTH, где у нас регистрируются и создаются пользователи
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
	log.Print("try to insert to chats")
	var chatId int64
	sql := "INSERT INTO chats DEFAULT VALUES RETURNING id"
	err := r.db.QueryRow(ctx, sql).Scan(&chatId)
	if err != nil {
		return 0, err
	}
	log.Printf("get chat_id: %d", chatId)
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

func (r *repo) SendMessage(ctx context.Context, msg model.Message) error {
	// Добавим проверку что chat_id существует и отправитель есть в этом чате
	if msg.ChatId <= 0 {
		return errors.New("chat_id cannot be empty")
	}

	if msg.FromUsername == "" || msg.Text == "" {
		return errors.New("from and text cannot be empty")
	}

	// Добавим валидацию, что пользователь есть в чате, делаю запрос без использования squirrel
	var isMember bool
	err := r.db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM chat_members WHERE chat_id = $1 AND username = $2)", msg.ChatId, msg.FromUsername).Scan(&isMember)
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
		Values(msg.ChatId, msg.FromUsername, msg.Text)
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
