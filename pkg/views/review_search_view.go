package views

import (
	"io"

	"github.com/aglide100/chicken_review_webserver/pkg/models"
)

type reviewSearchView struct {
	htmlctx   BaseHTMLContext
	reviews   []*models.Review
	CheckUser *models.User
}

func NewReviewSearchView(htmlctx BaseHTMLContext, reviews []*models.Review, CheckUser *models.User) View {
	return &reviewSearchView{htmlctx: htmlctx, reviews: reviews, CheckUser: CheckUser}
}

func (view reviewSearchView) ContentType() string {
	return "text/html"
}

func (view reviewSearchView) Render(w io.Writer) error {
	return view.htmlctx.RenderUsing(w, "ui/reviews/search.gohtml", view.reviews, view.CheckUser)
}
