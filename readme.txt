installing the gin web framework
go get github.com/gin-gonic/gin


the swagger site : https://goswagger.io/use/spec/meta.html
to generate swagger 
swagger generate spec -o ./swagger.json

swagger serve -F swagger ./swagger.json