const express = require('express')

  , http    = require('http')
  , path    = require('path')
  , db      = require('./models')

const basicAuth = require('express-basic-auth')

const app = express()

const csrfProtection = require('csurf')({ cookie: true })
const cookieParser = require('cookie-parser')

// all environments
app.set('port', process.env.PORT || 8000)
app.set('views', __dirname + '/views')
app.set('view engine', 'jade')
app.use(express.favicon())
app.use(express.logger('dev'))
app.use(express.json())
app.use(express.urlencoded())
app.use(express.methodOverride())
app.use(express.static(path.join(__dirname, 'public')))
app.use(cookieParser())
app.use(csrfProtection)

// development only
if ('development' === app.get('env')) {
  app.use(express.errorHandler())
}

function handleUnlisted(req, res)  {
  db.nodes.findAll({
    where: {
      listed: false,
      banned: false
    },
    order: [
      ['name', 'DESC']
    ]
  }).then((ns) => {
    res.render('unlisted', {nodes: ns.map((n) => n.toJSON())})
  })
}

app.use(basicAuth({
  authorizer: myAsyncAuthorizer,
  authorizeAsync: true,
  challenge: true,
  realm: 'phorebanserver'
}))

function myAsyncAuthorizer(username, password, cb) {
  if (username === 'phoreadmin' && password === process.env.PASSWORD)
    return cb(null, true)
  else
    return setTimeout(() => cb(null, false), 3000 + 200 * Math.random())
}

app.get('/', handleUnlisted);

app.get('/unlisted', handleUnlisted);

app.get('/items', (req, res) => {
  db.items.findAll().then((items) => {
    res.render('items', {items: items.map((i) => i.toJSON())})
  })
});

app.get('/banned', (req, res) => {
  db.nodes.findAll({
    where: {
      banned: true
    }
  }).then((ns) => {
    res.render('banned', {nodes: ns.map((n) => n.toJSON())})
  })
});

app.get('/list/:id', (req, res) => {
  db.nodes.find({
    where: {
      id: req.param("id")
    }
  }).then((n) => {
    n.listed = true
    return n.save()
  }).then(() => {
    res.redirect('/')
  })
})

app.get('/ban/:id', (req, res) => {
  db.nodes.find({
    where: {
      id: req.param("id")
    }
  }).then((n) => {
    n.banned = true
    return n.save()
  }).then(() => {
    res.redirect('/banned')
  })
})

app.get('/unlist/:id', (req, res) => {
  db.nodes.find({
    where: {
      id: req.param("id")
    }
  }).then((n) => {
    n.listed = false
    return n.save()
  }).then(() => {
    res.redirect('/banned')
  })
})

app.get('/unban/:id', (req, res) => {
  db.nodes.find({
    where: {
      id: req.param("id")
    }
  }).then((n) => {
    n.banned = false
    return n.save()
  }).then(() => {
    res.redirect('/')
  })
})

db
  .sequelize
  .sync()
  .complete(function(err) {
    if (err) {
      throw err
    } else {
      http.createServer(app).listen(app.get('port'), function(){
        console.log('Express server listening on port ' + app.get('port'))
      })
    }
  })
