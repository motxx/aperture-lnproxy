package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

const dbName = "db.json"

type DB struct {
	content *content
}

type content struct {
	Articles []*Article  `json:"articles"`
	Quotes   []*Quote    `json:"quotes"`
	Contents ContentsMap `json:"contents"`
}

func NewDB() (*DB, error) {
	// If there is no existing DB, create a new one. Otherwise, load the
	// existing one.
	file, err := os.Open(dbName)
	if errors.Is(err, os.ErrNotExist) {
		_, err = os.Create(dbName)
		if err != nil {
			return nil, err
		}

		return &DB{
			content: &content{},
		}, nil
	}
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	content := &content{}

	err = json.Unmarshal(byteValue, content)
	if err != nil {
		return nil, err
	}

	return &DB{
		content: content,
	}, nil
}

func (d *DB) Close() error {
	return d.writeContent()
}

func (d *DB) writeContent() error {
	b, err := json.MarshalIndent(d.content, " ", " ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dbName, b, 0644)
}

func (d *DB) AddContent(content *Content) (string, error) {
	if c, found := d.content.Contents[content.Id]; found {
		return "", fmt.Errorf("content with id %s already exists", c.Id)
	}

	d.content.Contents[content.Id] = content
	if err := d.writeContent(); err != nil {
		return "", err
	}
	return content.Id, nil
}

func (d *DB) UpdateContent(content *Content) (string, error) {
	if _, found := d.content.Contents[content.Id]; !found {
		return "", fmt.Errorf("no content with id %s", content.Id)
	}

	d.content.Contents[content.Id] = content
	if err := d.writeContent(); err != nil {
		return "", err
	}
	return content.Id, nil
}

func (d *DB) RemoveContent(id string) (string, error) {
	if _, found := d.content.Contents[id]; !found {
		return "", fmt.Errorf("no content with id %s", id)
	}

	delete(d.content.Contents, id)
	if err := d.writeContent(); err != nil {
		return "", err
	}
	return id, nil
}

func (d *DB) GetContent(id string) (*Content, error) {
	if c, ok := d.content.Contents[id]; ok {
		return c, nil
	}
	return nil, fmt.Errorf("no content with id %s", id)
}
