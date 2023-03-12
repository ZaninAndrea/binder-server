package rest

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ZaninAndrea/binder-server/internal/mongo"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/exp/slices"
)

type AuthenticationToken struct {
	jwt.RegisteredClaims
	Kind string `json:"kind,omitempty"`
}

func (t AuthenticationToken) Valid() error {
	now := time.Now()

	if !t.VerifyExpiresAt(now, false) {
		return ErrTokenExpired
	}
	if !t.VerifyNotBefore(now, false) {
		return ErrTokenNotValidYet
	}

	return nil
}

// SignToken generates a JWT string with a payload containing the specified
// kind and subject, the iat field is set to the current timestamp and the
// token is signed using the passed secret and the HS512 algorithm
func SignToken(kind string, subject string, secret []byte) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, AuthenticationToken{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
			Subject:  subject,
		},
		Kind: kind,
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// Authenticated returns a Gin handler function that checks whether
// the request contains an authentication token of the correct kind.
// The middleware expects the "Authorization" header to have already
// been parsed by the ParseAuthorizationHeader(...) middleware
func Authenticated(allowedTokenKinds []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check that an authorization token was parsed
		rawAuth, ok := c.Get("auth")
		if !ok {
			c.String(http.StatusBadRequest, `This route requires the authentication header to be set`)
			c.Abort()
			return
		}

		// Check that the token kind is allowed
		auth, ok := rawAuth.(AuthenticationToken)
		if !ok {
			c.String(http.StatusInternalServerError, `Failed to parse the authentication header`)
			c.Abort()
			return
		} else if !slices.Contains(allowedTokenKinds, auth.Kind) {
			c.String(
				http.StatusUnauthorized,
				"The passed token cannot be used on this route",
			)
			c.Abort()
			return
		}
	}
}

func GetAuthenticatedUser(c *gin.Context, db *mongo.Database) (bool, error, mongo.User) {
	id := c.MustGet("authSubject").(primitive.ObjectID)

	var user mongo.User
	exists, err := db.Users.FindByIdIfExists(id, &user)

	return exists, err, user
}

var ErrTokenExpired error = fmt.Errorf("the token is expired")
var ErrTokenNotValidYet error = fmt.Errorf("the token is not valid yet")

// ParseAuthorizationHeader returns a Gin handler function that checks
// that the "authorization" header, if present, contains a valid JWT token
// and stores the parsed JWT claims in Gin's request context: the whole
// token is stored in the "auth" field and the subject is stored in "authSubject"
func ParseAuthorizationHeader(jwtSecret []byte) gin.HandlerFunc {
	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS512"}))
	keyFunc := func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	}

	tokenValidationMiddleware := func(c *gin.Context) {
		rawAuth := c.GetHeader("authorization")
		if rawAuth == "" {
			return
		}

		if !strings.HasPrefix(rawAuth, "Bearer ") {
			c.String(http.StatusBadRequest, `The authorization header should have the "Bearer " prefix`)
			c.Abort()
			return
		}

		// Decode and verify the signature of the token
		token, err := parser.ParseWithClaims(rawAuth[7:], &AuthenticationToken{}, keyFunc)
		if err != nil {
			c.String(http.StatusBadRequest, `Invalid authorization header`)
			restLogger.Print(err)
			c.Abort()
			return
		}

		// Store the token content in the context
		claims, ok := token.Claims.(*AuthenticationToken)
		if !ok || !token.Valid {
			c.String(http.StatusBadRequest, `Invalid authorization header`)
			c.Abort()
			return
		}
		c.Set("auth", *claims)

		// Set the authSubject context variable if present in the token
		if claims.Subject != "" {
			subject, err := primitive.ObjectIDFromHex(claims.Subject)
			if err != nil {
				c.String(http.StatusBadRequest, `Invalid subject contained in the authorization token`)
				c.Abort()
				return
			}

			c.Set("authSubject", subject)
		}
	}

	return tokenValidationMiddleware
}
