package main

import (
	"flag"
	"log"
	"net/http"
)

const html = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Simple CORS</title>
</head>
<body>
<h1>Preflight CORS</h1>
<div id="output"></div>
<script>
    document.addEventListener('DOMContentLoaded', function (){
        fetch("http://localhost:4000/v1/tokens/authentication", {
            method: "POST",
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({
                email: 'abc@gmail.com',
                password: '1234567890'
            })
        }).then(
            function (response){
                response.text().then(function (text){
                    document.getElementById("output").innerHTML = text;
                })
            },
            function (err){
                document.getElementById("output").innerHTML = err;
                console.log(err.error)
            }
        )
    })
</script>

</body>
</html>
`

func main() {
	//Make the server configurable at runtime via command-line flag
	addr := flag.String("addr", ":9000", "Server address")
	flag.Parse()

	log.Printf("starting server on %d", addr)

	//Start HTTP Server
	err := http.ListenAndServe(*addr, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(html))
	}))
	log.Fatal(err)
}
