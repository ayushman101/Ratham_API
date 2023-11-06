package auth

import(
	"strings"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
)

var key string ="fejofjeaje335931jfjj3o"

//Authentication Function

func AuthToken(tokenString string) (string,error){

	ss := strings.Split(tokenString," ");

	if ss[0]!="Bearer"{
		return "",errors.New("No Bearer")
	}

	tokenString=ss[1]

    	if tokenString == "" {
        	//w.WriteHeader(http.StatusUnauthorized)
        	return "",errors.New("No token")
	}

    	// Validate the JWT token.
    	claims, err := ValidateJWT(tokenString, key)
    	if err != nil {
        	//w.WriteHeader(http.StatusUnauthorized)
		return "",fmt.Errorf("Invalid token: %w",err)
    	}

	fmt.Println(claims["id"])

	return fmt.Sprintf("%v",claims["id"]),nil

}


//sign a JWT token

func SignJWT(id interface{}) (string,error){
	token:=jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":id,
	})

	tokenString, err:= token.SignedString([]byte(key))

	if err!=nil{
		//log.Fatal(err)
		return "", fmt.Errorf("Error while signing token: %w",err)
	}

	return tokenString, nil
}

//Validate a token

func ValidateJWT(tokenString string, signingKey string) (jwt.MapClaims,error){

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
	// Don't forget to validate the alg is what you expect:
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
	}

	// signingKey is a []byte containing your secret, e.g. []byte("my_secret_key")
	return []byte(signingKey), nil
})
	
	if err!=nil{
		return nil, fmt.Errorf("invalid token %w",err)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid{
    		return claims, nil
	}

	return nil, fmt.Errorf("invalid token")
}
