const express = require('express');
const { get } = require('https');
const app = express()
const path = require('path'); 
const port = 3000

app.use(express.static('resources'))

const AUTHCODE_EXPIRATION = 5; // in minutes

const staticRelyingParty = {
    id: "3913614092307123",
    name: "Buy & Sell Guitars",
    domain: "http://localhost:3000/",
}

const relyingParties = {
    "3913614092307123": staticRelyingParty
}

const staticUser = {
    id: "4814028462",
    email: "bob@gmail.com",
    username: "a",
    password: "b",
    authCode: {
        code: "",
        timeSinceCreated: null,
    },
    permissions: [],
}

const users = {
    "4814028462": staticUser
}

function getUserbyID(id) {
    return users[id];
}

function getReylingPartybyID(id) {
    return relyingParties[id];
}

// auth codes are valid for 5min after issue
function validAuthCode(id, code) {
    // check user exists
    const user = getUserbyID(id);
    if (user === null) {
        return false
    }
    // check authcode was created
    const created = user.authCode.timeSinceCreated;
    const min = Math.abs(Math.round((((Date.now() - created) % 86400000) % 3600000) / 60000));
    return min < AUTHCODE_EXPIRATION;
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
        if (user.password == password && user.username == username) {
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
    } catch (error) {
        res.send("Invalid format")
    }
    if (!authRequest.scope.includes("openid")) {
        res.send("Invalid format: openid not in scope")
    }

    if (!checkAuthRequest(authRequest)) {
        res.send("Invalid format: ")
    }

    res.sendFile(path.join(__dirname, 'resources', 'login.html'))
});

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
    // check user's credentials
    const user = getUser(username, password);
    if (user != null) {
        // generate auth code
        user.authCode.timeSinceCreated = Date.now();
        user.authCode.code = randStr(20);
        res.json({success: true, code: user.authCode.code, userID: user.id})
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
        code = req.query.username;
    } catch (error) {
        res.json({success: false, reason: "Malformed request"})
        return
    }
    
    if (!validAuthCode(userID, code)) {
        res.json({success: false, reason: "Expired auth code"})
        return
    }

    res.sendFile(path.join(__dirname, 'resources', 'permissions.html'))
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
    // check that user exists
    const user = getUserbyID(userID)
    if (user == null) {
        res.json({success: false, redirect_uri: authRequest.redirect_uri, reason: "User does not exist"})
        return
    }
    // create temporary permissions code
    const curTime = Date.now();
    const perm = {
        code: randStr(20),
        created: curTime,
        min_expires: "3", // expires in 3 mins
        scope: authRequest.scope,
    }
    user.permissions.push(perm)
    res.json({success: true, redirect_uri: authRequest.redirect_uri, code: perm.code})
})

app.listen(port, () => {
    console.log(`Provider listening at http://localhost:${port}`)
})