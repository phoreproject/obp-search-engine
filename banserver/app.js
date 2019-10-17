'use strict';
const express = require('express'),
    http = require('http'),
    path = require('path'),
    db = require('./models'),
    request = require('request'),
    gmail = require('./gmailApi');

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
db.moderators.belongsToMany(db.nodes, {
    through: 'moderatorIdsPerItem',
    foreignKey: 'moderatorID',
    targetKey: 'id',
    otherKey: 'peerID'
});
db.nodes.belongsToMany(db.moderators, {
    through: 'moderatorIdsPerItem',
    foreignKey: 'peerID',
    targetKey: 'id',
    otherKey: 'moderatorID'
});

// development only
if ('development' === app.get('env')) {
    var errorHandler = require('errorhandler');
    app.use(errorHandler());
}

function handleUnlisted(req, res) {
    db.nodes.findAll({
        where: {
            listed: false,
            blocked: false,
            listingCount: {[db.sequelize.Op.ne]: 0},
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
    } else {
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


class RpcStatusChecker {
    constructor() {
        this.timeOfLastRPCcheck = 0;
        this.timeOfLastChainzCheck = 0;
        this.timeOfLastEmailSent = 0;

        this.rpcError = '';
        this.rpcLastBlock = 0;
        this.chainzError = '';
        this.chainzLastBlock = 0;
    }

    init() {
        setInterval(() => {
            request.post("https://rpc2.phore.io/rpc",
                {
                    json: {"jsonrpc": "2.0", "method": 'getblockcount', "params": [], "id": 1}
                },
                (err, resp, body) => {
                    if (err) {
                        this.rpcError = err.toString();
                        res.render('statistics', {tags: tags.map((tag) => tag.toJSON()), err: err});
                    } else if (resp.statusCode !== 200) {
                        this.rpcError = 'RPC returns status code ' + resp.statusCode;
                    } else {
                        this.rpcError= '';
                        this.rpcLastBlock = body.result;
                        this.timeOfLastRPCcheck = Date.now();
                    }
                    this.check_error();
                });
        }, 60 * 1000);


        setInterval(() => {
            request.post("https://chainz.cryptoid.info/phr/api.dws?q=getblockcount", (err, resp, body) => {
                if (err) {
                    this.chainzError = "chainz.cryptoid.info returns error " + err;
                } else if (resp.statusCode !== 200) {
                    this.chainzError = "chainz.cryptoid.info returns status code " + err;
                } else {
                    this.chainzError = '';
                    this.chainzLastBlock = parseInt(body, 10);
                    this.timeOfLastChainzCheck = Date.now();
                }
                this.check_error();
            });
        }, 60 * 1000);

        return this;
    }

    can_send_email() {
        return Date.now() - this.timeOfLastEmailSent > 15 * 60 * 1000;
    }

    check_error() {
        if ((this.rpcError !== '' || this.chainzError !== '') && this.can_send_email()) {
            let errMsg = '';
            if (this.rpcError !== '') {
                errMsg += 'An error occurred in RPC: ' + this.rpcError + '\n';
            } else {
                errMsg += 'RPC works correctly, but '
            }

            if (this.chainzError !== '') {
                errMsg += 'An error occured in Chainz: ' + this.chainzError + '\n';
            } else {
                errMsg += 'Chainz works correctly.'
            }

            gmail.sendEmail(gmail.createEmail("An error in ban manager occurred", errMsg, 'anchaj@phore.io'));
            this.timeOfLastEmailSent = Date.now();
        }
    }
}


const STATUS_CHECKER = new RpcStatusChecker().init();

app.get('/statistics', async (req, res) => {
    const tags = await db.nodes.findAll({
        group: ['userAgent'],
        attributes: ['userAgent', [db.sequelize.fn('COUNT', 'userAgent'), 'userAgentCnt']],
        order: [
            ['userAgent', 'DESC']
        ]
    });

    res.render('statistics', {
        tags: tags.map((tag) => tag.toJSON()),
        rpcBlockCount: STATUS_CHECKER.rpcLastBlock,
        chainzBlockCount: STATUS_CHECKER.chainzLastBlock,
        diff: STATUS_CHECKER.chainzLastBlock - STATUS_CHECKER.rpcLastBlock
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

        if (moderator == null && value === true) { // moderator doesn't exists yet, but we want to add one
            await db.moderators.create(
                {
                    id: req.params['id'],
                    isVerified: value,
                    type: 'standard'
                },
                {
                    transaction: transaction,
                })
        } else if (moderator != null) { // moderator exists
            let nodes = await db.nodes.findAll({
                include: [
                    {
                        model: db.moderators,
                        where: {
                            id: req.params['id']
                        }
                    }
                ],
                transaction: transaction,
            });

            if (value) {
                for (let i = 0; i < nodes.length; i++) {
                    await nodes[i].update({
                        moderator: true,
                        verifiedModerator: true,
                    }, {
                        transaction: transaction,
                    });
                }
            } else {
                for (let i = 0; i < nodes.length; i++) {
                    let mods = await db.moderators.findAll({
                        include: [
                            {
                                model: db.nodes,
                                where: {
                                    id: nodes[i].id
                                }
                            }
                        ],
                        where: {isVerified: true},
                        transaction: transaction
                    });

                    if (mods.length === 0) { // no more verified moderators for that node
                        nodes[i].update({
                            verifiedModerator: false
                        }, {
                            transaction: transaction
                        })
                    }
                }
            }

            await moderator.update({
                isVerified: value
            }, {
                transaction: transaction
            });
        }

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
