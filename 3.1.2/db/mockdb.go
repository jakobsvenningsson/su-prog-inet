package db

type MockDB struct {
	posts []Post
}

func NewMock() *MockDB {
	return &MockDB{posts: []Post{}}
}

func (mdb *MockDB) CreatePost(p Post) {
	mdb.posts = append(mdb.posts, p)
}

func (mdb *MockDB) GetPosts() []Post {
	return mdb.posts
}
