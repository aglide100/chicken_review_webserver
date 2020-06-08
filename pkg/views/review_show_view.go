package views

import (
	"io"

	"github.com/aglide100/chicken_review_webserver/pkg/models"
)

type reviewShowView struct {
	htmlctx BaseHTMLContext
	review  *models.Review
}

func NewReviewShowView(htmlctx BaseHTMLContext, review *models.Review) View {
	return &reviewShowView{htmlctx: htmlctx, review: review}
}

func (view reviewShowView) ContentType() string {
	return "text/html"
}

func (view reviewShowView) Render(w io.Writer) error {
	return view.htmlctx.RenderUsing(w, "ui/reviews/show.gohtml", view.review)
}
