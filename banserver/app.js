'use strict';
const express = require('express'),
    http = require('http'),
    path = require('path'),
    db = require('./models'),
    request = require('request');

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

// tables dependencies
db.moderators.belongsToMany(db.nodes, {through: 'moderatorIdsPerItem', foreignKey: 'moderatorID', targetKey: 'id', otherKey: 'peerID'});
db.nodes.belongsToMany(db.moderators, {through: 'moderatorIdsPerItem', foreignKey: 'peerID', targetKey: 'id', otherKey: 'moderatorID'});

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

app.get('/statistics', async (req, res) => {
    const tags = await db.nodes.findAll({
        group: ['userAgent'],
        attributes: ['userAgent', [db.sequelize.fn('COUNT', 'userAgent'), 'userAgentCnt']],
        order: [
            ['userAgent', 'DESC']
        ]
    });
    request.post("https://rpc.phore.io/rpc",
        {
            json: {"jsonrpc": "2.0", "method": 'getblockcount', "params": [], "id": 1}
        },
        (err, resp, body) => {


            if (err) {
                res.render('statistics', {tags: tags.map((tag) => tag.toJSON()), err: err});
            }
            else if (resp.statusCode !== 200) {
                res.render('statistics', {tags: tags.map((tag) => tag.toJSON()), err: 'RPC returns status code ' + resp.statusCode});
            }
            else {
                request.post("https://chainz.cryptoid.info/phr/api.dws?q=getblockcount", (errChainz, respChainz, bodyChainz) => {
                    if (errChainz) {
                        res.render('statistics', {tags: tags.map((tag) => tag.toJSON()), err: "RPC works (best block is"
                                + body.result + "), but chainz.cryptoid.info returns error " + err});
                    }
                    else if(respChainz.statusCode !== 200) {
                        res.render('statistics', {tags: tags.map((tag) => tag.toJSON()), err: "RPC works (best block is"
                                + body.result + "), but chainz.cryptoid.info returns status code " + err});
                    }
                    else {
                        res.render('statistics', {tags: tags.map((tag) => tag.toJSON()),
                            rpcBlockCount: body.result, chainzBlockCount: parseInt(bodyChainz, 10),
                            diff: parseInt(bodyChainz, 10) - body.result});
                    }
                });
            }
        });
});

app.get('/list/:id', (req, res) => {
    db.nodes.find({
        where: {
            id: req.params['id']
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
            id: req.params['id']
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
            id: req.params['id']
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
            id: req.params['id']
        }
    }).then((n) => {
        n.blocked = false;
        return n.save();
    }).then(() => {
        res.redirect('/');
    });
});

async function setIsVerified(req, res, value) {
    db.sequelize.transaction({}, async (transaction) => {
        let moderator = await db.moderators.find({
            where: {
                id: req.params['id']
            },
            transaction: transaction,
        });
        moderator.isVerified = value;
        await moderator.save();

        let nodes = await db.nodes.findAll({
            include:[
                {
                    model: db.moderators,
                }
            ],
            transaction: transaction,
        });

        //TODO update nodes

        res.redirect('/moderators');
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
