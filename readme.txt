installing the gin web framework
go get github.com/gin-gonic/gin


the swagger site : https://goswagger.io/use/spec/meta.html
to generate swagger 
swagger generate spec -o ./swagger.json

swagger serve -F swagger ./swagger.json


data entry code block:

	recipes = make([]Recipe, 0)
	file, _ := ioutil.ReadFile("recipes.json")
	err = json.Unmarshal(file, &recipes)
	if err != nil {
		log.Fatal(err)
	}
	var list []interface{}
	for _, recipe := range recipes {
		list = append(list, recipe)
	}
	insertedManyResult, err := collection.InsertMany(context.TODO(), list)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Inserted recipes ", len(insertedManyResult.InsertedIDs))


runnnig redis docker.

docker run -d -v $PWD/redis.conf:/usr/local/etc/redis/redis.conf --name redis -p 6379:6379 redis:6.0

go get github.com/go-redis/redis/v8


to check the log of redis docker:

docker exec -it container-id bash

for redis command line:

type : redis-cli

to check if data exists on redis:

EXISTS recipes

the above command will return 1 if data exists on redis.

redis insights is a gui tool for the redis docker.

docker run -d --name redis-insights --link redis -p 8001:8001 redislabs/redisinsight

after running the redis insight docker access the webpage at http://localhost:8001/

1. agree to the license agreement.
2. I already have a redis database
Host: redis
Port: 6379
database name : local
 add redis database


 to create a 16 digit secret key:

 openssl rand -base64 16

 simple api keys can be hacked  by man in the middle attack
 https://snyk.io/learn/man-in-the-middle-attack/


 jwt implementation:

 go get github.com/dgrijalva/jwt-go


 "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InZpamV5YXNoIiwiZXhwIjoxNjU4OTIxODAyfQ.lV3uKrT_Qi5ZiuMZQH5GfhdS3ArGp3v4nXJt2vzVTY0",


 session management:

 go get github.com/gin-contrib/sessions