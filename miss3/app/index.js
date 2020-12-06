const express = require('express');
const app = express()
const fs = require("fs");
const path = require('path'); 
const fetch = require('node-fetch');
const port = 8081;
const https = require('https');
const privateKey  = fs.readFileSync(path.join(__dirname, 'certs/app.key'), 'utf8');
const certificate = fs.readFileSync(path.join(__dirname, 'certs/app.cert'), 'utf8');
const credentials = {key: privateKey, cert: certificate};

// NOTE: TLS is only used here to encrypt traffic.
// official certs cannot be verified. Therefore, this 
// demo will trust all certs as long as they are provided.
// So cert attacks will not be allowed.
process.env['NODE_TLS_REJECT_UNAUTHORIZED'] = '0';

app.use(express.static('views'))
app.use(express.json());
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'ejs');

const CLIENT_SECRET = "mysupersecretkey";

app.get('/', (req, res) => {
    res.render('index');
}) 

app.get('/authenticating', (req, res) => {
    res.render('auth');
})

app.get('/account', (req, res) => {
    let params;
    try {
        params = {
            code: req.query.code
        }
    } catch(e) {
        res.redirect(`https://localhost:${port}/openid`);
        return
    }
    if (params.code === undefined) {
        res.redirect(`https://localhost:${port}/openid`);
        return
    }
    let uri = new URL("https://localhost:3000/token")
    fetch(uri, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
            'Authorization': 'Basic ' + CLIENT_SECRET,
            'grant_type': "authorization_code",
            'code': params.code,
            'redirect_uri':`https://localhost:${port}/token`
        },
        rejectUnauthorized: false
    }).then(res => res.json())
    .then(data => {
        // exchange token for userInfo
        console.log(data)
        let uri = new URL("https://localhost:3000/userinfo")
        uri.search = new URLSearchParams({
            id_token: data.id_token
        }).toString();
        fetch(uri, {
            method: 'POST'
        }).then(res => res.json())
        .then(data => {
            console.log(data)
            res.render('account', {
                FOUND_USERNAME: data.username,
                FOUND_EMAIL: data.email
            });
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
    console.log(`App listening at https://localhost:${port}`)
})