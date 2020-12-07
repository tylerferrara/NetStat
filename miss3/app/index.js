const express = require('express');
const app = express()
const fs = require("fs");
const path = require('path'); 
const fetch = require('node-fetch');
const https = require('https');
const privateKey  = fs.readFileSync(path.join(__dirname, 'certs/app.key'), 'utf8');
const certificate = fs.readFileSync(path.join(__dirname, 'certs/app.pem'), 'utf8');
const credentials = {key: privateKey, cert: certificate};

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
let port = 80;
let provPort = 80;
let appURI = "https://10.21.19.2";
let providerURI = "https://10.21.19.1";

if (DEV_ENV == "true") {
    port = 8081;
    provPort = 3000;
    providerURI = "https://localhost";
    appURI = "https://localhost";
}

app.use(express.static('views'))
app.use(express.json());
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'ejs');

const CLIENT_SECRET = "mysupersecretkey";

app.get('/', (req, res) => {
    res.render('index', {
        CONNECT_URI: `${appURI}:${port}/authenticating`
    });
}) 

app.get('/authenticating', (req, res) => {
    res.render('auth', {
        APP_URI: `${appURI}:${port}`,
        PROV_URI: `${providerURI}:${provPort}`
    });
})

app.get('/account', (req, res) => {
    let params;
    try {
        params = {
            code: req.query.code
        }
    } catch(e) {
        res.redirect(`${appURI}:${port}/`);
        return
    }
    if (params.code === undefined) {
        res.redirect(`${appURI}:${port}/`);
        return
    }
    let uri = new URL(`${providerURI}:${provPort}/token`)
    fetch(uri, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
            'Authorization': 'Basic ' + CLIENT_SECRET,
            'grant_type': "authorization_code",
            'code': params.code,
            'redirect_uri':`${appURI}:${port}/account`
        },
        rejectUnauthorized: false
    }).then(res => res.json())
    .then(data => {
        if (data.error != undefined) {
            res.redirect(`${appURI}:${port}`)
            return
        }
        // exchange token for userInfo
        console.log(data)
        let uri = new URL(`${providerURI}:${provPort}/userinfo`)
        uri.search = new URLSearchParams({
            id_token: data.id_token
        }).toString();
        fetch(uri, {
            method: 'POST'
        }).then(res => res.json())
        .then(data => {
            console.log(data)
            if (data.error == undefined) {
                res.render('account', {
                    FOUND_USERNAME: data.username,
                    FOUND_EMAIL: data.email
                });
                return
            } else {
                res.redirect(`${appURI}:${port}`)
                return
            }
        }).catch(e => {
            console.log(e);
            res.send("Failed to obtain account info")
        })
    }).catch(e => {
        console.log(e);
        res.send("Failed to obtain JWT token")
    })
    
})

const httpsServer = https.createServer(credentials, app);

httpsServer.listen(port, () => {
    console.log(`App listening at ${appURI}:${port}`)
})