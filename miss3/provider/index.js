const express = require('express');
const jwt = require('jsonwebtoken');
const crypto = require("crypto-js");
const app = express();
const fs = require('fs');
const https = require('https');
const path = require('path'); 
const privateKey  = fs.readFileSync(path.join(__dirname, 'certs/server.key'), 'utf8');
const certificate = fs.readFileSync(path.join(__dirname, 'certs/server.pem'), 'utf8');
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

// For password encryption
const PASS_KEY = "-2mcbal3ubi2oacs,2ioghe,a;;caij"

const DEV_ENV = process.env.DEV_ENV;
let port = 80;
let appPort = 80;
let mrRobotPort = 80;
let providerURI = "https://10.21.19.1";
let appURI = "https://10.21.19.2";
let mrRobotURI = "https://10.21.19.4"

if (DEV_ENV == "true") {
    port = 3000;
    appPort = 8081
    mrRobotPort = 4000;
    mrRobotURI = "https://localhost";
    providerURI = "https://localhost";
    appURI = "https://localhost";
}

app.use(express.static('views'))
app.use(express.json());
app.set('views', path.join(__dirname, 'views'));
app.set('view engine', 'ejs');

const AUTHCODE_EXPIRATION = 10; // in minutes
const PROVIDER_SECRET = "aff92fnasflfk2fnv02hgms";

const relyingParties = {
    "3913614092307123": {
        id: "3913614092307123",
        secret: "mysupersecretkey",
        name: "Buy & Sell Guitars",
        domain: `${appURI}:${appPort}/`,
        sub: null,
    },
    "4923047502348671": {
        id: "4923047502348671",
        secret: "anotherwacksecret",
        name: "Mr.Robot",
        domain: `${mrRobotURI}:${mrRobotPort}/`,
        sub: null,
    }
}

const keyCodes = [];

const users = {
    "4814028462": {
        id: "4814028462",
        email: "kentclark@gmail.com",
        username: "superman",
        password: "U2FsdGVkX18Fs8iI/GYToZbkIX21ZtLwdlB88JmNpLw=", // hash of iFlyhigh
        authCode: {
            code: "",
            created: null,
            min_expires: AUTHCODE_EXPIRATION,
        },
        permissions: [],
    },
    "2720247213": {
        id: "2720247213",
        email: "handymany@gmail.com",
        username: "many",
        password: "U2FsdGVkX18mNThv4dQ+3/abtmNyaCMc34nCowWmkn4=", // hash of ifixit
        authCode: {
            code: "",
            created: null,
            min_expires: AUTHCODE_EXPIRATION,
        },
        permissions: [],
    }
}

function getUserbyID(id) {
    return users[id];
}

function getReylingPartybyID(id) {
    return relyingParties[id];
}

// auth codes are valid for 5min after issue
function validAuthCode(authCode) {
    if (authCode === null) {
        return false
    }
    // check authcode was created
    const min = Math.abs(Math.round((((Date.now() - authCode.created) % 86400000) % 3600000) / 60000));
    return min < authCode.min_expires;
}

function randStr(length) {
    var result           = '';
    var characters       = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
    var charactersLength = characters.length;
    for ( var i = 0; i < length; i++ ) {
       result += characters.charAt(Math.floor(Math.random() * charactersLength));
    }
    return result;
}

function getUser(username, password) {
    const keys = Object.keys(users);
    for (let i = 0; i < keys.length; i++) {
        const user = users[keys[i]];
        const p = crypto.AES.decrypt(user.password, PASS_KEY).toString(crypto.enc.Utf8)
        if (p == password && user.username == username) {
            return user;
        }
    }
    return null;
}

function checkAuthRequest(req) {
    const party = getReylingPartybyID(req.client_id)
    return party != null;
}

app.get('/authenticate', (req, res) => {
    let authRequest;
    try {
        authRequest = {
            responseType: req.query.response_type,
            scope: req.query.scope,
            client_id: req.query.client_id,
            redirect_uri: req.query.redirect_uri,
        }
    } catch(error) {
        res.send("Invalid format")
    }
    if (!authRequest.scope.includes("openid")) {
        res.send("Invalid format: openid not in scope")
    }

    if (!checkAuthRequest(authRequest)) {
        res.send("Invalid format: ")
    }
    res.render('login', {
        PROV_URI: `${providerURI}:${port}`
    });
});

app.post('/validatecookie', (req, res) => {
    let cookie
    try {
        cookie = req.query.cookie;
    } catch (error) {
        res.json({success: false, reason: "Malformed request"})
        return
    }
    // decode JWT
    let decoded;
    try {
        decoded = jwt.verify(cookie, PROVIDER_SECRET);
    } catch(e) {
        console.log(e);
        res.json({success: false, reason: "JWT token couldn't be verified"})
        return
    }
    // validate
    const user = users[decoded.userID]
    if (user == undefined || user == null) {
        res.json({success: false, reason: "No matching user"})
        return
    }
    if (!validAuthCode(user.authCode)) {
        res.json({success: false, reason: "Expired auth code"})
        return
    }
    res.json({success: true, userID: decoded.userID, code: decoded.code})
})

app.post('/authenticate', (req, res) => {
    let authRequest;
    let username;
    let password;
    try {
        authRequest = {
            responseType: req.query.response_type,
            scope: req.query.scope,
            client_id: req.query.client_id,
            redirect_uri: req.query.redirect_uri,
        }
        username = req.query.username;
        password = req.query.password;
    } catch (error) {
        res.json({success: false, reason: "Malformed request"})
        return
    }
    if (username == undefined || password == undefined) {
        res.json({success: false, reason: "Credentials not found"})
    }
    // check user's credentials
    const user = getUser(username, password);
    if (user != null) {
        // generate auth code
        user.authCode.created = Date.now();
        user.authCode.code = randStr(20);   // set auth code
        // create cookie for future logins
        const cookie = jwt.sign({
            code: user.authCode.code,
            userID: user.id,
        }, PROVIDER_SECRET);
        // send response
        res.json({success: true, code: user.authCode.code, userID: user.id, cookie: cookie})
    } else {
        res.json({success: false, reason: "Invalid credentials"})
    }
});

app.post('/providername', (req, res) => {
    const id = req.query.client_id;
    const party = getReylingPartybyID(id);
    if (party == null) {
        res.json({success: false})
    } else {
        res.json({success: true, client_name: party.name})
    }
})

app.get('/permissions', (req, res) => {
    let authRequest, code, userID;
    try {
        authRequest = {
            responseType: req.query.response_type,
            scope: req.query.scope,
            client_id: req.query.client_id,
            redirect_uri: req.query.redirect_uri,
        }
        userID = req.query.userID;
        code = req.query.code;
    } catch (error) {
        res.json({success: false, reason: "Malformed request"})
        return
    }
    // didn't verify with temp code
    const user = getUserbyID(userID);
    if (user == undefined) {
        res.json({success: false, reason: "User not found"})
        return
    }
    if (!validAuthCode(user.authCode)) {
        res.json({success: false, reason: "Expired auth code"})
        return
    }

    res.render('permissions', {
        PROV_URI: `${providerURI}:${port}`
    });
});

app.post('/permissions', (req, res) => {
    let authRequest, accept, userID;
    try {
        authRequest = {
            responseType: req.query.response_type,
            scope: req.query.scope,
            client_id: req.query.client_id,
            redirect_uri: req.query.redirect_uri,
        }
        userID = req.query.userID;
        accept = req.query.accept;
    } catch (error) {
        res.json({success: false, redirect_uri: authRequest.redirect_uri, reason: "Malformed request"})
        return
    }
    // did the user accept?
    if (!accept) {
        res.json({success: false, redirect_uri: authRequest.redirect_uri, reason: "User does not accept"})
        return
    }
    // check that user exists
    const user = getUserbyID(userID)
    if (user == null) {
        res.json({success: false, redirect_uri: authRequest.redirect_uri, reason: "User does not exist"})
        return
    }
    // create temporary permissions code
    const curTime = Date.now();
    const perm = {
        userID: userID,
        clientID: authRequest.client_id,
        code: randStr(20),
        created: curTime,
        min_expires: 3, // expires in 3 mins
        scope: authRequest.scope,
    }
    user.permissions.push(perm)
    keyCodes.push(perm)
    res.json({success: true, redirect_uri: authRequest.redirect_uri, code: perm.code})
})

app.post('/token', (req, res) => {
    let auth, grant_type, redirect_uri, code;
    try {
        auth = req.header('Authorization')
        grant_type = req.header('grant_type');
        redirect_uri = req.header('redirect_uri');
        code = req.header('code');
    } catch(error) {
        res.status = 400;
        res.json({"error": "invalid_request"})
        return
    }
    if (auth == undefined || grant_type == undefined || redirect_uri == undefined || code == undefined) {
        res.status = 400;
        res.json({"error": "invalid_request"})
        return
    }
    // only accept authorization_code type
    if (grant_type != "authorization_code") {
        res.status = 400;
        res.json({"error": "invalid_request"})
        return
    }
    // check for authorization secret in providers
    let matchingParty = null;
    let keys = Object.keys(relyingParties)
    for (let i = 0; i < keys.length; i++) {
        const party = relyingParties[keys[i]];
        if ('Basic ' + party.secret === auth) {
            matchingParty = party;
            break;
        }
    }
    if (matchingParty === null) {
        res.status = 400;
        res.json({"error": "invalid_request"})
        return
    }
    // look for keyCodes
    let tempCode = null;
    keyCodes.forEach((v) => {
        // valid expiration
        if (v.clientID == matchingParty.id && validAuthCode(v)) {
            tempCode = v;
        }
    })
    // validate request
    if (tempCode === null || code != tempCode.code) {
        res.status = 400;
        res.json({"error": "invalid_request"})
        return
    }
    // create JWT Token
    const curTime = new Date(Date.now());
    const expires = new Date(curTime.getTime() + 30*60000); // 30min from now
    matchingParty.sub = randStr(25);
    const contents = {
        iss: `${providerURI}:${port}`,
        sub: matchingParty.sub,
        aud: matchingParty.clientID,
        exp: expires.getTime(),
        iat: curTime.getTime()
    }
    const token = jwt.sign(contents, PROVIDER_SECRET);
    res.json({expires_in: contents.exp, id_token: token})
})

app.post('/userinfo', (req, res) => {
    let id_token;
    try {
        id_token = req.query.id_token;
    } catch(e) {
        console.log(e);
        res.json({error: "invalid_request", reason: "id_token not provided in request"})
        return
    }
    if (id_token == undefined) {
        res.json({error: "invalid_request"})
        return
    }
    // decode JWT
    let decoded;
    try {
        decoded = jwt.verify(id_token, PROVIDER_SECRET);
    } catch(e) {
        console.log(e);
        res.json({error: "invalid_request", reason: "JWT token couldn't be verified"})
        return
    }

    // check expiration date
    const curTime = new Date(Date.now());
    if (decoded.exp <= curTime.getTime()) {
        res.json({error: "invalid_request"})
        return
    }
    // identify the relying party
    let matchingParty = null;
    const keys = Object.keys(relyingParties);
    for (let i = 0; i < keys.length; i++) {
        const party = relyingParties[keys[i]];
        if (party.sub === decoded.sub) {
            matchingParty = party;
            break;
        }
    }
    if (matchingParty === null) {
        // found no relying party to match JWT sub identifier
        res.json({error: "invalid_request", reason: "no relying party found"})
        return
    }
    // find user who gave permissions to relying party
    let foundUser = null;
    const ukeys = Object.keys(users);
    for (let i = 0; i < ukeys.length; i++) {
        const user = users[ukeys[i]];
        for (let j = 0; j < user.permissions.length; j++) {
            let p = user.permissions[j];
            if (p.clientID === matchingParty.id) {
                foundUser = user;
                break;
            }
        }
    }
    if (foundUser === null) {
        // no user permissions to give
        res.json({error: "invalid_request", reason: "no user found"})
        return
    }
    res.json({email: foundUser.email, username: foundUser.username})
})

const httpsServer = https.createServer(credentials, app);

httpsServer.listen(port, () => {
    console.log(`Provider listening at ${providerURI}:${port}`)
})