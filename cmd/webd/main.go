package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/aglide100/chicken_review_webserver/pkg/api"
	"github.com/aglide100/chicken_review_webserver/pkg/controllers"
	"github.com/aglide100/chicken_review_webserver/pkg/models"
	"github.com/aglide100/chicken_review_webserver/pkg/router"
	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
	"github.com/markbates/goth/providers/kakao"
	"github.com/markbates/goth/providers/naver"

	"github.com/aglide100/chicken_review_webserver/pkg/db"
)

var (
	//listenHttpPort  = os.Getenv("LISTEN_HTTP_PORT")
	listenAddr      = os.Getenv("LISTEN_ADDR")
	listenPort      = os.Getenv("LISTEN_HTTPS_PORT")
	hostHTTPport    = os.Getenv("HOST_HTTP_PORT")
	hostHTTPSport   = os.Getenv("HOST_HTTPS_PORT")
	tlsCertFilepath = os.Getenv("TLS_CERT_FILEPATH")
	tlsKeyFilepath  = os.Getenv("TLS_KEY_FILEPATH")
	dbAddr          = os.Getenv("DB_ADDR")
	dbPort          = os.Getenv("DB_PORT")
	dbUser          = os.Getenv("DB_USER")
	dbPassword      = os.Getenv("DB_PASSWORD")
	dbName          = os.Getenv("DB_NAME")
	callbackAddr    = os.Getenv("CALLBACK_ADDR")
	KakaoMaps       = os.Getenv("KAKAO_MAPS_API_KEYS")
	sessionKey      = []byte(os.Getenv("SESSIONKEY"))
)

func main() {
	// return to https connection
	//go http.ListenAndServe(":"+hostHTTPport, http.HandlerFunc(redirect))

	if os.Getenv("SESSIONKEY") == "" {
		sessionKey = securecookie.GenerateRandomKey(32)
		log.Printf("Can't find session key! please input session key!!!")
	}

	if err := realMain(); err != nil {
		log.Fatal(err)
	}
}

var store = sessions.NewCookieStore(sessionKey)

/*
func redirect(resp http.ResponseWriter, req *http.Request) {
	// for Using html5 function, so redirect to https connection
	var target string
	if hostHTTPSport == "443" {
		target = "https://" + req.Host + "/reviews"
	} else {
		target = "https://" + req.Host + ":" + hostHTTPSport + "/reviews"
	}
	if len(req.URL.RawQuery) > 0 {
		target += "?" + req.URL.RawQuery
	}
	log.Printf("redirect to: %s", target)
	http.Redirect(resp, req, target, 301)
}
*/

func realMain() error {
	log.Printf("start realMain")

	log.Printf("ListenAddr : %v:%v, DBAddr : %v:%v, DBUser : %v, DBPWD : %v", listenAddr, listenPort, dbAddr, dbPort, dbUser, dbPassword)

	/* Using goth */
	goth.UseProviders(
		naver.New(os.Getenv("NAVER_KEY"), os.Getenv("NAVER_SECRET"), callbackAddr+"/auth/callback?provider=naver"),
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), callbackAddr+"/auth/callback?provider=google"),
		kakao.New(os.Getenv("KAKAO_KEY"), os.Getenv("KAKAO_SECRET"), callbackAddr+"/auth/callback?provider=kakao"),
	)

	sessions.NewSession(store, "session-name")
	gothic.Store = sessions.NewCookieStore(sessionKey)
	// add Api keys(GoogleMaps)
	APIKeys := &models.APIKeys{
		KakaoMaps: KakaoMaps,
		//GoogleMaps: GoogleMaps,
	}
	addr := net.JoinHostPort(listenAddr, listenPort)

	dbport, _ := strconv.Atoi(dbPort)
	myDB, err := db.ConnectDB(dbAddr, dbport, dbUser, dbPassword, dbName)
	if err != nil {
		return fmt.Errorf("connecting to DB: %v", err)
	}

	defaultCtrl := &controllers.DefaultController{}
	notFoundCtrl := &controllers.NotFoundController{}
	sessionCtrl := controllers.NewSessionController(store)
	reviewsCtrl := controllers.NewReviewController(myDB, APIKeys, sessionCtrl)
	ajaxCtrl := api.NewAjaxController(sessionCtrl)
	loginCtrl := controllers.NewLoginController(myDB, sessionCtrl)

	rtr := router.NewRouter(notFoundCtrl)

	rtr.AddRule("default", "GET", "^/$", defaultCtrl.ServeHTTP)

	rtr.AddRule("ajax", "GET", "^/ajax", ajaxCtrl.AjaxHandler)

	rtr.AddRule("login", "GET", "^/login/register_page$", loginCtrl.Register_Page)
	rtr.AddRule("login", "POST", "^/login/sign_up", loginCtrl.Register)

	rtr.AddRule("login", "GET", "^/login", loginCtrl.LoginCheck)
	rtr.AddRule("login", "GET", "^/logout", ajaxCtrl.SessionLogOut)
	rtr.AddRule("login", "POST", "^/login/log_In", loginCtrl.LogIn)
	rtr.AddRule("login", "GET", "^/login/log_Out", loginCtrl.LogOut)

	rtr.AddRule("login", "GET", "^/auth$", loginCtrl.AuthGoth)
	//rtr.AddRule("login", "GET", "^/auth[?]provider=[a-zA-Z]*$", loginCtrl.AuthGoth)
	rtr.AddRule("login", "GET", "^/auth/logout/([A-Za-z])$", loginCtrl.GothLogOut)
	//rtr.AddRule("login", "GET", "^/auth/callback[?]provider[=][a-zA-Z]*[&]*", loginCtrl.GothCallBack)
	rtr.AddRule("login", "GET", "^/auth/callback", loginCtrl.GothCallBack)

	rtr.AddRule("reviews", "GET", "^/reviews/?$", reviewsCtrl.List)
	rtr.AddRule("reviews", "GET", "^reviews/([A-Z]{1,3})-pagenumber=([0-9]+)$", reviewsCtrl.List)
	rtr.AddRule("reviews", "GET", "^/reviews/([0-9]+)$", reviewsCtrl.Get)

	rtr.AddRule("reviews", "GET", "^/reviews/create$", reviewsCtrl.Create)
	rtr.AddRule("reviews", "POST", "^/reviews/create/upload", reviewsCtrl.Save)

	rtr.AddRule("reviews", "GET", "^/update/([0-9]+)$", reviewsCtrl.Revise)
	rtr.AddRule("reviews", "POST", "^/reviews/update/upload/", reviewsCtrl.Update)

	rtr.AddRule("reviews", "GET", "^/delete/([0-9]+)$", reviewsCtrl.Delete)

	//rtr.AddRule("reviews", "GET", "^/reviews/search/", reviewsCtrl.Search)
	rtr.AddRule("reviews", "POST", "^/reviews/search/post", reviewsCtrl.Search)

	//rtr.AddRule("reviews", "GET", "^/img", reviewsCtrl.GetImage)
	//rtr.AddRule("reviews", "GET", "^/reviews/ui/img/([0-9]+)/[a-z0-9A-Z_+.-.\\s.-]+.(?i)(img|jpg|jpeg|png|gif)$", reviewsCtrl.GetImage)

	// URI ex) reviews/ui/img/1/0/1.jpeg
	rtr.AddRule("reviews", "GET", "^/reviews/ui/img/([0-9]+)/[a-z0-9_+.-]/[a-z0-9A-Z_+.-.\\s.-]+.(?i)(img|jpg|jpeg|png|gif)$", reviewsCtrl.GetImage)
	rtr.AddRule("reviews", "GET", "^/reviews/ui/logo/[a-zA-Z]*/[a-z0-9A-Z_+.-.\\s.-]+.(?i)(img|jpg|jpeg|png|gif)$", reviewsCtrl.GetImage)
	rtr.AddRule("reviews", "GET", "^/reviews/ui/err/no_image.png$", reviewsCtrl.GetImage)

	rtr.AddRule("reviews", "GET", "^/reviews/ui/css/.*", reviewsCtrl.GetScript)
	rtr.AddRule("reviews", "GET", "^/reviews/ui/js/.*", reviewsCtrl.GetScript)
	//rtr.AddRule("reviews", "GET", "^/reviews/ui/assets/", reviewsCtrl.GetAssets)

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("creating network listener: %v", err)
	}
	defer ln.Close()

	srv := http.Server{Handler: rtr}
	log.Printf("listening on address %q", ln.Addr().String())

	log.Printf("starting server at address %q", ln.Addr().String())

	err = srv.ServeTLS(ln, tlsCertFilepath, tlsKeyFilepath)
	//err = srv.Serve(ln)
	if err != nil {
		return fmt.Errorf("serving: %v", err)
	}

	return nil
}
