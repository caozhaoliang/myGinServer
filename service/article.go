package service

import (
	"context"
	"myGinServer/internal/store"
	"myGinServer/models/article"
)

type ArticleServer struct {
	dbStore store.DBStore
}

func NewArticleServer(db store.DBStore) *ArticleServer {
	return &ArticleServer{
		dbStore: db,
	}
}

func (a *ArticleServer) Channels(ctx context.Context) ([]article.Channel, error) {
	return a.dbStore.Channels(ctx)
}

func (a *ArticleServer) Articles(ctx context.Context, req *article.ArticlesRequest) (article.ArticlesResponse, error) {
	return a.dbStore.GetArticles(ctx, req)
}
