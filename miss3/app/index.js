const express = require('express');
const app = express()
const path = require('path'); 
const port = 8081

const STATIC_CLIENT_ID = "3913614092307123"

app.use(express.static('resources'))

app.get('/', (req, res) => {
    res.sendFile('index.html')
})

app.get('/openid', (req, res) => {
    let uri = new URL('http://localhost:3000/authenticate')
    uri.search = new URLSearchParams({
        response_code: "code",
        scope: "openid profile",
        client_id: STATIC_CLIENT_ID,
        redirect_uri: `http://localhost:${port}/account`
    }).toString();
    res.redirect(uri)
})

app.get('/account', (req, res) => {
    res.send("You made it to the account page!!!")
})

app.listen(port, () => {
    console.log(`App listening at http://localhost:${port}`)
})