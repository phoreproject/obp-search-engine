'use strict';
const express = require('express'),
    http = require('http'),
    path = require('path'),
    db = require('./models');

const basicAuth = require('express-basic-auth');

const app = express();

const csrfProtection = require('csurf')({cookie: true});
const cookieParser = require('cookie-parser');

// all environments
app.set('port', process.env.PORT || 8000);
app.set('views', __dirname + '/views');
app.set('view engine', 'pug');
// Middlewares
let favicon = require('serve-favicon');
let morgan = require('morgan');
let methodOverride = require('method-override');
app.use(favicon(path.join(__dirname, 'public', 'images', 'Phore_16_x_16.png')));
app.use(morgan('combined'));
app.use(express.json());
app.use(express.urlencoded());
app.use(methodOverride());
app.use(express.static(path.join(__dirname, 'public')));
app.use(cookieParser());
app.use(csrfProtection);

// development only
if ('development' === app.get('env')) {
    var errorHandler = require('errorhandler');
    app.use(errorHandler());
}

function handleUnlisted(req, res) {
    db.nodes.findAll({
        where: {
            listed: false,
            blocked: false
        },
        order: [
            ['name', 'DESC']
        ]
    }).then((ns) => {
        res.render('unlisted', {nodes: ns.map((n) => n.toJSON())});
    });
}

function myAsyncAuthorizer(username, password, cb) {
    if (username === 'phoreadmin' && password === process.env.PASSWORD) {
        return cb(null, true);
    }
    else {
        return setTimeout(() => cb(null, false), 3000 + 200 * Math.random());
    }
}

app.use(basicAuth({
    authorizer: myAsyncAuthorizer,
    authorizeAsync: true,
    challenge: true,
    realm: 'phorebanserver'
}));

app.get('/', handleUnlisted);

app.get('/unlisted', handleUnlisted);

app.get('/banned', (req, res) => {
    db.nodes.findAll({
        where: {
            blocked: true
        }
    }).then((ns) => {
        res.render('banned', {nodes: ns.map((n) => n.toJSON())});
    });
});

app.get('/moderators', (req, res) => {
    db.moderators.findAll().then((ns) => {
        res.render('moderators', {moderators: ns.map((n) => n.toJSON())});
    });
});

app.get('/list/:id', (req, res) => {
    db.nodes.find({
        where: {
            id: req.param('id')
        }
    }).then((n) => {
        n.listed = true;
        return n.save();
    }).then(() => {
        res.redirect('/');
    });
});

app.get('/unlist/:id', (req, res) => {
    db.nodes.find({
        where: {
            id: req.param('id')
        }
    }).then((n) => {
        n.listed = false;
        return n.save();
    }).then(() => {
        res.redirect('/banned');
    });
});

app.get('/ban/:id', (req, res) => {
    db.nodes.find({
        where: {
            id: req.param('id')
        }
    }).then((n) => {
        n.blocked = true;
        return n.save();
    }).then(() => {
        res.redirect('/banned');
    });
});

app.get('/unban/:id', (req, res) => {
    db.nodes.find({
        where: {
            id: req.param('id')
        }
    }).then((n) => {
        n.blocked = false;
        return n.save();
    }).then(() => {
        res.redirect('/');
    });
});

function setIsVerified(req, res, value) {
    db.moderators.find({
        where: {
            id: req.param('id')
        }
    }).then((mod) => {
        mod.isVerified = value;
        return mod.save();
    }).then(() => {
        res.redirect('/moderators')
    });
}

app.get('/verify/:id', (req, res) => {
    setIsVerified(req, res, true);
});

app.get('/unverify/:id', (req, res) => {
    setIsVerified(req, res, false);
});

db.sequelize.sync().then(function () {
    http.createServer(app).listen(app.get('port'), function () {
        console.log('Express server listening on port ' + app.get('port'));
    });
}, function (err) {
    throw(err);
});
