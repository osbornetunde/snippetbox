package main

import (
	"net/http"
	"net/url"
	"testing"
	"tundeosborne.snippetbox/internal/assert"
)

func TestPing(t *testing.T) {

	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	code, _, body := ts.get(t, "/ping")

	assert.Equal(t, code, http.StatusOK)
	assert.Equal(t, body, "OK")
}

func TestSnippetView(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	tests := []struct {
		name     string
		urlPath  string
		wantCode int
		wantBody string
	}{
		{
			name:     "Valid ID",
			urlPath:  "/snippet/view/1",
			wantCode: http.StatusOK,
			wantBody: "An old silent pond...",
		},
		{
			name:     "Non-existent ID",
			urlPath:  "/snippet/view/2",
			wantCode: http.StatusNotFound,
		}, {
			name:     "Negative ID",
			urlPath:  "/snippet/view/-1",
			wantCode: http.StatusNotFound,
		}, {
			name:     "Decimal ID",
			urlPath:  "/snippet/view/1.23",
			wantCode: http.StatusNotFound,
		}, {
			name:     "String ID",
			urlPath:  "/snippet/view/foo",
			wantCode: http.StatusNotFound,
		}, {
			name:     "Empty ID",
			urlPath:  "/snippet/view/",
			wantCode: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, _, body := ts.get(t, tt.urlPath)
			assert.Equal(t, code, tt.wantCode)
			if tt.wantBody != "" {
				assert.StringContains(t, body, tt.wantBody)
			}
		})
	}
}

func TestUserSignup(t *testing.T) {
	app := newTestApplication(t)

	ts := newTestServer(t, app.routes())
	defer ts.Close()

	_, _, body := ts.get(t, "/user/signup")
	validCSRFToken := extractCSRFToken(t, body)

	const (
		validName     = "Bob"
		validEmail    = "bob@example.com"
		validPassword = "validPa$$word"
		formTag       = "<form action='/user/signup' method='POST' novalidate>"
	)

	tests := []struct {
		name         string
		userName     string
		userEmail    string
		userPassword string
		csrfToken    string
		wantCode     int
		wantFormTag  string
	}{
		{
			name:         "Valid submission",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusSeeOther,
		}, {
			name:         "Invalid CSFR Token",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    "wrongToken",
			wantCode:     http.StatusBadRequest,
		}, {
			name:         "Empty Name",
			userName:     "",
			userEmail:    validEmail,
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		}, {
			name:         "Empty Email",
			userName:     validName,
			userEmail:    "",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		}, {
			name:         "Empty Password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		}, {
			name:         "Invalid Email",
			userName:     validName,
			userEmail:    "bob@example.",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		}, {
			name:         "Short Password",
			userName:     validName,
			userEmail:    validEmail,
			userPassword: "pa$$",
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		}, {
			name:         "Duplicate Email",
			userName:     validName,
			userEmail:    "dupe@example.com",
			userPassword: validPassword,
			csrfToken:    validCSRFToken,
			wantCode:     http.StatusUnprocessableEntity,
			wantFormTag:  formTag,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			form := url.Values{}
			form.Add("name", tt.userName)
			form.Add("email", tt.userEmail)
			form.Add("password", tt.userPassword)
			form.Add("csrf_token", tt.csrfToken)

			code, _, body := ts.postForm(t, "/user/signup", form)

			assert.Equal(t, code, tt.wantCode)
			if tt.wantFormTag != "" {
				assert.StringContains(t, body, tt.wantFormTag)
			}
		})
	}

	t.Logf("CSRF Token: %q", validCSRFToken)
}

func TestSnippetCreate(t *testing.T) {
	app := newTestApplication(t)
	ts := newTestServer(t, app.routes())
	defer ts.Close()

	t.Run("Unauthenticated", func(t *testing.T) {
		code, headers, _ := ts.get(t, "/snippet/create")

		assert.Equal(t, http.StatusSeeOther, code)
		assert.Equal(t, headers.Get("Location"), "/user/login")
	})

	t.Run("Authenticated", func(t *testing.T) {
		aboveHundredLengthTitle := "Lorem ipsum Nunc Mauris posuere blandit fermentum morbi Vivamus Aliquam nec vitae neque potenti ex purus dapibus feugiat eleifend amet Fusce Quisque mi dolor facilisis feugiat vitae Nulla Ut sem porta quis fermentum Pellentesque volutpat senectus ligula volutpat suscipit molestie suscipit lectus Ut varius a Etiam Donec erat Morbi Nam posuere id lobortis In Suspendisse elit auctor finibus Morbi ornare Aenean Vivamus Aliquam sem quis molestie pulvinar Aenean Aenean molestie vulputate efficitur Vestibulum sem Nam Phasellus Quisque amet Praesent Curabitur eu quis habitasse Donec maximus rutrum consectetur Sed facilisis Mauris Mauris libero Suspendisse tincidunt cursus Nunc Aenean Nam ultrices bibendum rhoncus"
		_, _, body := ts.get(t, "/user/login")
		validAuthCsrfToken := extractCSRFToken(t, body)

		const (
			validEmail    = "alice@example.com"
			validPassword = "pa55word"
			formTag       = "<form action='/snippet/create' method='POST'>"
		)

		tests := []struct {
			name            string
			userEmail       string
			userPassword    string
			signInCsrfToken string
			signInWantCode  int
			wantFormTag     string
			wantCode        int
			title           string
			content         string
			expires         string
			viewSnippetPath string
		}{
			{
				name:            "Authenticated and create snippet",
				userEmail:       validEmail,
				userPassword:    validPassword,
				signInCsrfToken: validAuthCsrfToken,
				signInWantCode:  http.StatusOK,
				wantFormTag:     formTag,
				wantCode:        http.StatusSeeOther,
				title:           "Hello World",
				content:         "Happy Birthday people",
				expires:         "365",
				viewSnippetPath: "/snippet/view/2",
			}, {
				name:            "Authenticated and fail to create snippet due empty title",
				userEmail:       validEmail,
				userPassword:    validPassword,
				signInCsrfToken: validAuthCsrfToken,
				signInWantCode:  http.StatusOK,
				wantFormTag:     formTag,
				wantCode:        http.StatusUnprocessableEntity,
				title:           "",
				content:         "Happy Birthday people",
				expires:         "365",
				viewSnippetPath: "",
			}, {
				name:            "Authenticated and fail to create snippet due empty content",
				userEmail:       validEmail,
				userPassword:    validPassword,
				signInCsrfToken: validAuthCsrfToken,
				signInWantCode:  http.StatusOK,
				wantFormTag:     formTag,
				wantCode:        http.StatusUnprocessableEntity,
				title:           "Hello World",
				content:         "",
				expires:         "365",
				viewSnippetPath: "",
			}, {
				name:            "Authenticated and fail to create snippet due empty expires",
				userEmail:       validEmail,
				userPassword:    validPassword,
				signInCsrfToken: validAuthCsrfToken,
				signInWantCode:  http.StatusOK,
				wantFormTag:     formTag,
				wantCode:        http.StatusUnprocessableEntity,
				title:           "Hello World",
				content:         "Happy Birthday people",
				expires:         "",
				viewSnippetPath: "",
			}, {
				name:            "Authenticated and fail to create snippet when expires has wrong value",
				userEmail:       validEmail,
				userPassword:    validPassword,
				signInCsrfToken: validAuthCsrfToken,
				signInWantCode:  http.StatusOK,
				wantFormTag:     formTag,
				wantCode:        http.StatusUnprocessableEntity,
				title:           "Hello World",
				content:         "Happy Birthday people",
				expires:         "4",
				viewSnippetPath: "",
			}, {
				name:            "Authenticated and fail to create snippet when expires has wrong value",
				userEmail:       validEmail,
				userPassword:    validPassword,
				signInCsrfToken: validAuthCsrfToken,
				signInWantCode:  http.StatusOK,
				wantFormTag:     formTag,
				wantCode:        http.StatusUnprocessableEntity,
				title:           aboveHundredLengthTitle,
				content:         "Happy Birthday people",
				expires:         "365",
				viewSnippetPath: "",
			}}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				form := url.Values{}
				form.Add("email", tt.userEmail)
				form.Add("password", tt.userPassword)
				form.Add("csrf_token", tt.signInCsrfToken)
				ts.postForm(t, "/user/login", form)

				code, _, body := ts.get(t, "/snippet/create")
				assert.Equal(t, code, tt.signInWantCode)
				assert.StringContains(t, body, tt.wantFormTag)

				newCsrfTokens := extractCSRFToken(t, body)
				forms := url.Values{}
				forms.Add("title", tt.title)
				forms.Add("content", tt.content)
				forms.Add("expires", tt.expires)
				forms.Add("csrf_token", newCsrfTokens)

				codes, headers, _ := ts.postForm(t, "/snippet/create", forms)
				assert.Equal(t, codes, tt.wantCode)
				assert.Equal(t, headers.Get("Location"), tt.viewSnippetPath)

			})
		}

	})
}
