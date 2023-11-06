package controllers

import(
	"time"
	"fmt"
	"net/http"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"encoding/json"
	"context"
	"log"
	"github.com/ayushman101/warden_go_mongo/models"
	//"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"github.com/ayushman101/warden_go_mongo/authenticate"
)




//for user queries
type UserController struct{
	Client *mongo.Client
}


func NewUserController(c *mongo.Client) *UserController{
	return &UserController{c}
}


//Register a User
func (uc UserController) CreateUser(w http.ResponseWriter, r *http.Request){


	var user models.User

	err:= json.NewDecoder(r.Body).Decode(&user)
	if err!=nil {
		fmt.Println(err)
		return 
	}
	
	user.ID=primitive.NewObjectID()

	if user.UserType!="Dean"{
		user.UserType="Student"
	}

	collection:= uc.Client.Database("Ratham_db").Collection("users")

	result,err:= collection.InsertOne(context.Background(),user)

	if err!=nil{
		 fmt.Println(err)
		 return
	}
	


	tok,err:= auth.SignJWT(result.InsertedID);
	
	if err!=nil{
		 log.Fatal(err)
	}

	json.NewEncoder(w).Encode(tok)

}


//List All users
func (uc UserController) Allusers(w http.ResponseWriter, r *http.Request){
	//Validating the User
	tokenString := r.Header.Get("Authorization")
	
	fmt.Println("inside All users")
	_,err:=auth.AuthToken(tokenString)

	if err!=nil{
		fmt.Printf("auth:",err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	collection := uc.Client.Database("Ratham_db").Collection("users")

	
	//second argument is of type Document which is defined in mongo-bson.
	cursor,err:= collection.Find(context.Background(),bson.D{{}})  //second argument is a query filer
								       // empty for returnig all users
	if err!=nil{
		fmt.Println("1",err)
		//log.Fatal(err)
		
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var users []models.User

	//iterate over the cursor to get all the users
	for cursor.Next(context.Background()){

		var user models.User

		err = cursor.Decode(&user)

		if err!=nil{
			fmt.Println(err)
	
			//log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		users=append(users,user)
	}


	//close cursor
	err= cursor.Close(context.Background())
	
	if err!=nil{
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//return users in json
	json.NewEncoder(w).Encode(users)

}


//login function

func (uc UserController) LoginUser(w http.ResponseWriter, r *http.Request){
	
	var user models.User
	
		
	bodybytes,err:= ioutil.ReadAll(r.Body)

	if err!=nil {
		fmt.Println("ioutil:",err)
		w.WriteHeader(http.StatusBadRequest)
		return 
	}

	err=json.Unmarshal(bodybytes, &user)

	if err!=nil{
		fmt.Println("Unmarshal:",err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}



	collection:= uc.Client.Database("Ratham_db").Collection("users")
	
	err= collection.FindOne(context.Background(), bson.M{"_id":user.ID,"email":user.Email,"password":user.Password}).Decode(&user)

	if err!=nil{
		fmt.Println(" finding user :",err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tok,err:= auth.SignJWT(user.ID);
	
	if err!=nil{
		 //log.Fatal(err)
		 fmt.Println("signJwt :",err)
		 w.WriteHeader(http.StatusInternalServerError)
		 return 
	}

	json.NewEncoder(w).Encode(tok)

}



func (uc UserController) CreateSession(w http.ResponseWriter, r *http.Request){
	tokenString := r.Header.Get("Authorization")
	
	fmt.Println("inside create session controller")
	
	id,err:=auth.AuthToken(tokenString)
	
	if err!=nil{
		fmt.Println("session: error while auth",err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//converting string type to valid bson ObjectID
	id1,err:=primitive.ObjectIDFromHex(id)


	//TODO: Change Session time to store next THU 11 AM 
	session:= models.DeanSession{
		ID: primitive.NewObjectID(),
		DeanId: id1,
		Status: "available",
		SessionTime: primitive.NewDateTimeFromTime(time.Now()),
		ExpiresAt: primitive.NewDateTimeFromTime(time.Now().Add(time.Second * 180)),
	}

	collection:= uc.Client.Database("Ratham_db").Collection("Dean_Sessions")

	collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
        	Keys: bson.M{
            		"expiresAt": 1,
        	},
        	Options: options.Index().SetExpireAfterSeconds(0),
    	})

	result,err:= collection.InsertOne(context.Background(),session)

	if err!=nil{
		 fmt.Println(err)
		 w.WriteHeader(http.StatusBadRequest)
		 return
	}
	
	json.NewEncoder(w).Encode(result.InsertedID)

}


//List Available Sessions of a Warden 
func (uc UserController) ListAvailableSessions(w http.ResponseWriter, r *http.Request){
	//Validating the User
	tokenString := r.Header.Get("Authorization")
	
	//fmt.Println("inside All users")
	_,err:=auth.AuthToken(tokenString)

	if err!=nil{
		fmt.Printf("auth:",err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var user models.User
	
	// unmarshalling the body of request
	bodybytes,err:= ioutil.ReadAll(r.Body)

	if err!=nil {
		fmt.Println("ioutil:",err)
		w.WriteHeader(http.StatusBadRequest)
		return 
	}

	err=json.Unmarshal(bodybytes, &user)

	if err!=nil{
		fmt.Println("Unmarshal:",err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Finding the existence of the Dean	
	collection:= uc.Client.Database("Ratham_db").Collection("users")

	err= collection.FindOne(context.Background(), bson.M{"_id":user.ID}).Decode(&user)
	
	if err!=nil{
		fmt.Println(" finding user :",err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if user.UserType!="Dean" {

		fmt.Println("Not Dean",err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	//Findig all the sessions
	collection = uc.Client.Database("Ratham_db").Collection("Dean_Sessions")
	

	if err!=nil{
		fmt.Println(" finding user :",err)
		w.WriteHeader(http.StatusNotFound)
		return
	}


	cursor,err:= collection.Find(context.Background(),bson.M{"dean_id":user.ID, "status":"available"})  //second argument is a query filer
								       // empty for returnig all users
	if err!=nil{
		fmt.Println("1",err)
		//log.Fatal(err)
		
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var sessions []models.DeanSession

	//iterate over the cursor to get all the users
	for cursor.Next(context.Background()){

		var session models.DeanSession

		err = cursor.Decode(&session)

		if err!=nil{
			fmt.Println(err)
	
			//log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		sessions=append(sessions,session)
	}

	fmt.Println(sessions)

	//close cursor
	err= cursor.Close(context.Background())
	
	if err!=nil{
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//return users in json
	json.NewEncoder(w).Encode(sessions)

}


//Pending Sessions:

func (uc UserController) PendingSessions(w http.ResponseWriter, r *http.Request){
	//Validating the User
	tokenString := r.Header.Get("Authorization")
	
	//fmt.Println("inside All users")
	id,err:=auth.AuthToken(tokenString)

	if err!=nil{
		fmt.Printf("auth:",err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	id1,err:=primitive.ObjectIDFromHex(id)
	
	if err!=nil{
		fmt.Println("invalid id:",err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	
	//Findig all the sessions
	collection := uc.Client.Database("Ratham_db").Collection("Dean_Sessions")
	

	cursor,err:= collection.Find(context.Background(),bson.M{"dean_id":id1, "status":"pending"})  //second argument is a query filer
								       // empty for returnig all users
	if err!=nil{
		fmt.Println("1",err)
		//log.Fatal(err)
		
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var sessions []models.DeanSession

	//iterate over the cursor to get all the users
	for cursor.Next(context.Background()){

		var session models.DeanSession

		err = cursor.Decode(&session)

		if err!=nil{
			fmt.Println(err)
	
			//log.Fatal(err)
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		sessions=append(sessions,session)
	}


	//close cursor
	err= cursor.Close(context.Background())
	
	if err!=nil{
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//return users in json

	//TODO: Set the header Content-Type to json
	json.NewEncoder(w).Encode(sessions)
	
}



//Booking session function

func (uc UserController) BookSession(w http.ResponseWriter, r *http.Request){
	//Validating the User
	tokenString := r.Header.Get("Authorization")
	
	//fmt.Println("inside All users")
	id,err:=auth.AuthToken(tokenString)
	if err!=nil{
		fmt.Printf("auth:",err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}


	id1,err:=primitive.ObjectIDFromHex(id)
	if err!=nil{
		fmt.Printf("Invalid id: ",err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}


	var dean models.DeanSession
	
	// unmarshalling the body of request
	bodybytes,err:= ioutil.ReadAll(r.Body)

	if err!=nil {
		fmt.Println("ioutil:",err)
		w.WriteHeader(http.StatusBadRequest)
		return 
	}

	err=json.Unmarshal(bodybytes, &dean)

	if err!=nil{
		fmt.Println("Unmarshal:",err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	//Findig all the sessions
	collection := uc.Client.Database("Ratham_db").Collection("Dean_Sessions")
	

	fmt.Println(dean.DeanId)

	update:=bson.M{
		
		"$set":bson.M{"booker_id":id1, "status": "pending",},
	}



	result, err:=collection.UpdateOne(context.Background(),bson.M{"dean_id":dean.DeanId, "status":"available", "session_time":dean.SessionTime},update)

	if err!=nil{
		fmt.Println("While updating: ",err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	 if result.MatchedCount == 0 {
        fmt.Println("No documents matched the filter criteria.")
    } else if result.ModifiedCount == 0 {
        fmt.Println("The document was found, but nothing was updated.")
    } else {
        fmt.Println("The document was found and updated.")
	w.WriteHeader(http.StatusOK)
    }
	
	json.NewEncoder(w).Encode(result)
	

}

