package main

import (
	"context"
	"time"
	"encoding/json"
	"log"
	"net/http"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"go.mongodb.org/mongo-driver/v2/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"strconv"
)

 type User struct{
	Fname    string `json:"fname"`
	Lname    string `json:"lname"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Question struct {
	Category      string   `json:"category"`
	Difficulty    string   `json:"difficulty"`
	Question      string   `json:"question"`
	Answers       []string `json:"answers"`
	CorrectAnswer string   `json:"correctAnswer"`
}


func hashingPasswords(password string) (string, error){
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(bytes), err
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Println("The Method I'm recieving is:", r.Method)
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST,GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	w.Write([]byte(`{"message":"Signup received"}`))
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST,GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	var data User
	log.Println("I'm here!")
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
	log.Println("The data:", &data)
	log.Println(err)
	return
	}
	hashedPassword, err := hashingPasswords(data.Password)
	if err != nil {
    http.Error(w, "Error hashing password", http.StatusInternalServerError)
    return
	}
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
  	opts := options.Client().ApplyURI("mongodb+srv://HoneyLemon:q5XnISPim8iWS5ho@cluster1.62zgzjg.mongodb.net/?appName=Cluster1").SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(opts)
	if err != nil {
    panic(err)
  }
  defer func() {
    if err = client.Disconnect(context.TODO()); err != nil {
      panic(err)
    }
  }()
  if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
    panic(err)
  }
	log.Println("Pinged your deployment. You successfully connected to MongoDB!")
	
	db := client.Database("quizapp")
    usersCollection := db.Collection("users")
	newUser := User{
        Fname:    data.Fname,
		Lname:    data.Lname,
		Username: data.Username,
		Email:    data.Email,
		Password: hashedPassword,
    }
    result, err := usersCollection.InsertOne(context.TODO(), newUser)
    if err != nil {
        log.Fatal(err)
    }
	log.Println("Inserted user:", result)
	log.Println("Stored password:", newUser.Password)
	log.Println("Length:", len(newUser.Password))
}

func getUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	// Connect to Mongo
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI("mongodb+srv://HoneyLemon:q5XnISPim8iWS5ho@cluster1.62zgzjg.mongodb.net/?appName=Cluster1").SetServerAPIOptions(serverAPI)
	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("quizapp")
	usersCollection := db.Collection("users")

	// Decode request body to get username
	var reqData struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&reqData); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var user User
	err = usersCollection.FindOne(context.TODO(), bson.M{"username": reqData.Username}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Don't send the password to frontend
	user.Password = ""

	json.NewEncoder(w).Encode(user)
}

func searchUser(w http.ResponseWriter, r *http.Request) {
	//log.Println("Hello?")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST,GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
    serverAPI := options.ServerAPI(options.ServerAPIVersion1)
    clientOptions := options.Client().ApplyURI("mongodb+srv://HoneyLemon:q5XnISPim8iWS5ho@cluster1.62zgzjg.mongodb.net/?appName=Cluster1").SetServerAPIOptions(serverAPI)
    client, err := mongo.Connect(clientOptions)
	//log.Println("I'm here!")
    if err != nil {
        log.Fatal(err)
    }
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    defer client.Disconnect(ctx)
	//log.Println("Connected to MongoDB!")
    if err := client.Ping(ctx, readpref.Primary()); err != nil {
        log.Fatal(err)
    }
    db := client.Database("quizapp")
    usersCollection := db.Collection("users")
    cursor, err := usersCollection.Find(ctx, bson.M{})
    if err != nil {
        log.Fatal(err)
    }
    defer cursor.Close(ctx)
	var data User
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
	log.Println(err)
	return
	}
	var user User
	err = usersCollection.FindOne(ctx, bson.M{
		"username": data.Username,
	}).Decode(&user)

	if err != nil {
		log.Println("User not found")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// DEBUG (you should see bcrypt hash here)
	log.Println("DB password length:", len(user.Password))
	log.Println("DB password:", user.Password)

	// 🔐 BCRYPT COMPARE
	err = bcrypt.CompareHashAndPassword(
		[]byte(user.Password), // ✅ HASH FROM DB
		[]byte(data.Password), // ✅ PLAIN PASSWORD
	)
	if err != nil {
		log.Println("Password mismatch")
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	log.Println("✅ Login successful")
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
	w.WriteHeader(http.StatusOK)
	return
	}

	if r.Method != http.MethodPut {
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Println("Decode error:", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	log.Printf("Received update: %+v\n", user)
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI("mongodb+srv://HoneyLemon:q5XnISPim8iWS5ho@cluster1.62zgzjg.mongodb.net/?appName=Cluster1").SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(clientOptions)
	if err != nil {
		http.Error(w, "DB connection failed", 500)
		return
	}
	defer client.Disconnect(context.TODO())

	collection := client.Database("quizapp").Collection("users")
	update := bson.M{
		"$set": bson.M{
			"fname":    user.Fname,
			"lname":    user.Lname,
			"email":    user.Email,
			"username": user.Username,
			"password": user.Password,
		},
	}

	if user.Password != "" {
		hashed, _ := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		update["$set"].(bson.M)["password"] = string(hashed)
	}

	result, err := collection.UpdateOne(
    context.TODO(),
    bson.M{"username": user.Username},
    update,
	)

	if err != nil {
	    log.Println("MongoDB update error:", err)
	    http.Error(w, "Update failed", 500)
	    return
	}

	if result.MatchedCount == 0 {
		http.Error(w, "User not found", 404)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "User updated successfully",
	})
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
	w.WriteHeader(http.StatusOK)
	return
	}

	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Connect to Mongo
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().
		ApplyURI("mongodb+srv://HoneyLemon:q5XnISPim8iWS5ho@cluster1.62zgzjg.mongodb.net/?appName=Cluster1").
		SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(opts)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(context.TODO())

	db := client.Database("quizapp")
	usersCollection := db.Collection("users")

	// Get username from request body
	var data struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	result, err := usersCollection.DeleteOne(context.TODO(), bson.M{"username": data.Username})
	if err != nil {
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	if result.DeletedCount == 0 {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": "User deleted successfully"})
}

func getCategory(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
    serverAPI := options.ServerAPI(options.ServerAPIVersion1)
    clientOptions := options.Client().ApplyURI("mongodb+srv://HoneyLemon:q5XnISPim8iWS5ho@cluster1.62zgzjg.mongodb.net/?appName=Cluster1").SetServerAPIOptions(serverAPI)
    client, err := mongo.Connect(clientOptions)
	if err != nil {
        log.Fatal(err)
    }
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    defer client.Disconnect(ctx)
    if err := client.Ping(ctx, readpref.Primary()); err != nil {
        log.Fatal(err)
    }
    db := client.Database("Quiz")
    quizCollection := db.Collection("Questions")
	category := r.URL.Query().Get("category")
	difficulty := r.URL.Query().Get("difficulty")
	amountStr := r.URL.Query().Get("amount")
	amount, err := strconv.Atoi(amountStr)
	if err != nil || amount <= 0 {
		http.Error(w, "invalid amount", http.StatusBadRequest)
		return
	}
	filter := bson.D{
		{"category", category},
		{"difficulty", difficulty},
	}
	opts := options.Find().SetLimit(int64(amount))

	// Query
	cursor, err := quizCollection.Find(ctx, filter, opts)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(ctx)
	var results []Question
	for cursor.Next(ctx) {
		var q Question
		if err := cursor.Decode(&q); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		results = append(results, q)
	}
	json.NewEncoder(w).Encode(results)
	w.Header().Set("Content-Type", "application/json")
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/signup", createUser)
	http.HandleFunc("/login", searchUser)
	http.HandleFunc("/getuser", getUser)
	http.HandleFunc("/update", updateUser)
	http.HandleFunc("/delete", deleteUser)
	http.HandleFunc("/questions", getCategory)
	log.Println("Listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
