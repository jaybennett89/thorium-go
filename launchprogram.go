package main

import (
	"crypto/md5"
	"crypto/rsa"
	"encoding/json"
	"hussain/github.com/jaybennett89/thorium-go/requests"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-martini/martini"
)
import "os/exec"
import "fmt"

import "redis"
import "bytes"

import "crypto/rand"

import "database/sql"
import _ "github.com/lib/pq"

//For now these field are from the sql table of games,
//proper struct should be
/*
type ContractInformation struct {
	OfferedBy string
	TimeRemaining int
	Bid int
	Ask int
}
*/
type ContractInformation struct {
	game_id     int
	map_name    string
	max_players int
	is_verified bool
}

func Check(prog string) int {
	acceptedList := []string{"boltactiongame", "test"}
	for i := 0; i < len(acceptedList); i++ {
		if acceptedList[i] == prog {
			return 1
		}
	}
	return 0

}

//returns cmd struct of program
func Execute(prog string) *exec.Cmd {
	cmd := exec.Command("./" + prog)
	e := cmd.Start()
	if e != nil {
		fmt.Println("Error runninng program ", e)
	}
	return cmd

}

func RedisPush(rawtokenstring string) int {
	spec := redis.DefaultSpec().Password("go-redis")
	client, e := redis.NewSynchClientWithSpec(spec)
	if e != nil {
		fmt.Println("error creating client for: ", e)
	}
	defer client.Quit()
	//pidString := strconv.Itoa(cmdI.Process.Pid)
	decryptedToken, err := jwt.Parse(rawtokenstring, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	//fmt.Printf("token strings\nRaw: [%s]\nHeader: [%s]\nSignature: [%s]\n", decryptedToken.Raw, decryptedToken.Header, decryptedToken.Signature)

	//check if no error and valid token
	if err == nil && decryptedToken.Valid {
		fmt.Println("token is valid and not expired")

	} else {
		fmt.Println("Not valid: ", err)
		return 0
	}

	userID := decryptedToken.Claims["id"].(string)
	//fmt.Println("redis func userid: " + userID + "\n redis func raw: " + decryptedToken.Raw)
	var buf bytes.Buffer
	buf.Write([]byte(decryptedToken.Raw))
	e = client.Hset("clients/token", userID, buf.Bytes())
	//to retrieve token in redis-cli, do hget clients/tokens legacy
	if e != nil {
		fmt.Println("error writing to list")
		return 0
	}
	return 1
}

func PostGresQueryIDS() []int {
	db, err := sql.Open("postgres", "user=thoriumnet password=thoriumtest dbname=thoriumnet host=localhost")
	if err != nil {
		fmt.Println("err: ", err)
	}
	var game_id int
	game_ids := make([]int, 100)
	rows, err := db.Query("SELECT * FROM games;")
	if err != nil {
		fmt.Println("err2: ", err)
	}
	defer rows.Close()
	for rows.Next() {
		err := rows.Scan(&game_id)
		if err != nil {
			fmt.Println("err3: ", err)
		}
		for index, _ := range game_ids {
			if game_ids[index] == 0 {
				game_ids[index] = game_id
				break
			}
		}
	}
	return game_ids
}

var secretKey string

func main() {
	//rand.Seed(time.Now().UnixNano())
	//var portArg = flag.Int("p", rand.Intn(65000-10000)+10000, "specifies port, default is random int between 10000-65000")
	//var mapArg = flag.String("m", "default map  value", "description of map")
	//flag.Parse()
	//fmt.Println(strconv.Itoa(*portArg))
	//fmt.Println(*mapArg)
	//rand.Seed = 1
	//processL := make([]*exec.Cmd, 100)
	//currentGames := PostGresQueryIDS()
	//for _, value := range currentGames {
	//if value != 0 {
	//fmt.Println(value)
	//}
	//}
	m := martini.Classic()
	secretKey = "superdupersecretkey"
	/*m.Post("/launch/:name",  func(params martini.Params) string {
		e := Check(params["name"])
		if e==1 {
			cmdInfo := Execute(params["name"])
			for i:=0; i<len(processL); i++ {
				if processL[i]==nil {
					processL[i]=cmdInfo
					//suc := RedisPush(cmdInfo)
					RedisPush(cmdInfo)
					break
				}
			}
			//fmt.Println(processL)
			return "launching " + params["name"] + "with pid " + strconv.Itoa(cmdInfo.Process.Pid)
		} else {
			return "not accepted"
		}
	})
	*/
	//m.Get("/games", gameServerInfo)
	m.Post("/client/login", handleClientLogin)
	m.Post("/client/afterlogin", handleAfterLogin)
	m.Run()
	//	err := cmd.Wait()
	//	fmt.Println(err)
	//	fmt.Println(cmd.Path)
	//	fmt.Println(cmd.Process.Pid)
}

func handleAfterLogin(httpReq *http.Request) (int, string) {
	var req request.Test
	decoder := json.NewDecoder(httpReq.Body)
	err := decoder.Decode(&req)
	if err != nil {
		fmt.Println("error decoding token request")
		return 500, "Internal Server Error"
	}

	//need to return the secret key to the parse to verify the token
	decryptedToken, err := jwt.Parse(req.Token, func(token *jwt.Token) (interface{}, error) {
		return []byte(secretKey), nil
	})
	fmt.Printf("token strings\nRaw: [%s]\nHeader: [%s]\nSignature: [%s]\n", decryptedToken.Raw, decryptedToken.Header, decryptedToken.Signature)

	//check if no error and valid token
	if err == nil && decryptedToken.Valid {
		fmt.Println("token is valid and not expired")
		//wrote this to check expirey but .Valid already does that
		/*
			expiredTime := decryptedToken.Claims["exp"].(float64)
			if float64(time.Now().Unix()) > expiredTime {
				return 500, "token expired get out of here"
			} else {
				return 200, "token is valid and not expired"
			}

			fmt.Println(decryptedToken.Claims)
		*/
	} else {
		fmt.Println("Not valid: ", err)
		return 500, "Internal Server Error"
	}

	return 200, "ok"
}

func handleClientLogin(httpReq *http.Request) (int, string) {
	decoder := json.NewDecoder(httpReq.Body)
	var req request.Authentication
	err := decoder.Decode(&req)
	if err != nil {
		//logerr("Error decoding authentication request")
		fmt.Println("error with json: ", err)
		return 500, "Internal Server Error"
	}
	//need to check if username && password are correct.

	/*
		if req.Username == database username && req.Password == database password
		then we can start to generate the token
	*/

	//create new token
	token := jwt.New(jwt.SigningMethodRS256)

	//secret key is used for signing and verifying token
	//secretKey := "superdupersecretkey"

	//generate private/public key for encrypting/decrypting token claims (if we need to)
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		fmt.Println("Could not generate key: ", err)
		return 500, "Internal Server Error"
	}
	//get the public key from the private key
	publicKey := &privateKey.PublicKey

	//need these vars for encryption/decryption
	md5hash := md5.New()
	label := []byte("")

	//actual encryption
	encryptedUsername, err := rsa.EncryptOAEP(md5hash, rand.Reader, publicKey, []byte(req.Username), label)
	if err != nil {
		fmt.Println("error encrypting: ", err)
	}
	//set the UserID value to the encrypted username, not sure if needed.
	token.Claims["id"] = req.Username
	//2 minute expiery
	token.Claims["exp"] = time.Now().Add(time.Minute * 2).Unix()
	//decrypt to check if encryption  worked properly

	decryptedUsername, err := rsa.DecryptOAEP(md5hash, rand.Reader, privateKey, encryptedUsername, label)
	if err != nil {
		fmt.Println("error decrypting: ", err)
	}

	fmt.Printf("decrypted [%x] to \n[%s]\n", token.Claims["UserID"], decryptedUsername)
	/*
		encrypter, err := NewEncrypter(RSA_OAEP,A128GCM, publicKey)
		if err != nil {
			fmt.Println("Algorithm not supported")
			return 500, "Internal Server Error"
		}
	*/
	//fmt.Println(*publicKey)

	//need to sign the token with something, for now its a random string
	tokenString, err := token.SignedString([]byte(secretKey))
	if err != nil {
		fmt.Println("error getting signed key: ", err)
		return 500, "Internal Server Error"
	}

	success := RedisPush(tokenString)
	if success == 1 {
		fmt.Println("pushed to redis")
	} else {
		fmt.Println("could not push to redis")
	}

	//return 200, tokenString + "\n"
	fmt.Println("Token String: ", tokenString)
	return 200, tokenString
}

/*
func gameServerInfo() string {
	db, err := sql.Open("postgres", "user=thoriumnet password=thoriumtest dbname=thoriumnet host=localhost")
	if err != nil {
		fmt.Println("database conn err: ", err)
		//return 500, err
	}
	//var tx *sql.Tx
	//tx, e = db.Begin()

	//store contract information
	var info ContractInformation
	//get game id
	rows, err := db.Query("SELECT * FROM games")
	if err != nil {
		fmt.Println("error: ", err)
	}
	defer rows.Close()
	//scan row by row
	for rows.Next() {
		//must scan all variables
		err := rows.Scan(&info.game_id, &info.map_name, &info.max_players, &info.is_verified)
		if err != nil {
			fmt.Println("error scanning row: ", err)
		}
		fmt.Println("id: ", info.game_id, "map: ", info.map_name, "max_players: ", info.max_players, "verified: ", info.is_verified)

	}

	return "finished"
}*/
