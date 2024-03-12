package db

type Article struct {
	Title   string `json:"title"`
	Author  string `json:"author"`
	Content string `json:"content"`
}

type Quote struct {
	Author  string `json:"author"`
	Content string `json:"content"`
	Price   int64  `json:"price"`
}

type Content struct {
	Title    string `json:"title"`
	Author   string `json:"author"`
	Filepath string `json:"filepath"`
	Price    int64  `json:"price"`
}

type ContentsMap map[string]*Content
