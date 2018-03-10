const express = require("express")
const app = express()
const Sequelize = require("sequelize")

const sequelize = new Sequelize(process.env.DATABASE_URI || "mysql://root@localhost:3306/obpsearch", {omitNull: true, logging: false });

const Item = sequelize.import("./models/item")

app.get("/logo.png", (req, res) => {
    res.sendFile("./static/logo.png")
})

const config = require('./config')

app.get("/", (req, res) => {
    res.send()
})

app.get('/search/listings', (req, res) => {
    const options = {}  
    const page = req.query.p || 1
    const ps = Math.min(req.query.ps || 20, 100)
    const nsfw = req.query.nsfw || false
    const orderBy = req.query.sortBy || "RELEVANCE"
    
    options.limit = ps
    options.offset = ps * (page - 1)
    options.where = {}
    options.where.nsfw = nsfw

    if (req.query.rating) {
        options.where.rating = {
            $gt: {
                5: 4.75,
                4: 4,
                3: 3,
                2: 2,
                1: 1
            }[Number(req.query.rating)]
        }
    }
    options.order = []
    
    if (orderBy.startsWith("PRICE")) {
        options.order[0] = "price"
    } else if (orderBy.startsWith("RATING")) {
        options.order[0] = "rating"
    }
    if (orderBy.endsWith("DESC")) {
        options.order[1] = "DESC"
    } else if (orderBy.endsWith("ASC")) {
        options.order[1] = "ASC"
    }
    Item.findAndCountAll(options).then((out) => {
        const result = Object.assign(config, {
            results: {
                total: out.count,
                morePages: out.count > ps,
                results: []
            }
        })
        for (r of out.rows) {
            thumbnails = r.thumbnail.split(",")
            result.results.results.push({
                type: "listing",
                relationships: {
                    peerID: r.owner,
                    handle: "",
                    avatarHashes: {
                        tiny: thumbnails[0], // TODO: this is wrong
                        small: thumbnails[1],
                        medium: thumbnails[2]
                    },
                    moderator: []
                },
                data: {
                    hash: r.hash,
                    slug: r.slug,
                    title: r.title,
                    tags: r.tags.split(","),
                    contractType: r.contractType,
                    description: r.description,
                    thumbnail: {
                        tiny: thumbnails[0],
                        small: thumbnails[1],
                        medium: thumbnails[2]
                    },
                    language: r.language,
                    price: {
                        amount: r.priceAmount,
                        currencyCode: r.priceCurrency
                    },
                    nsfw: r.nsfw,
                    categories: r.categories.split(",")
                }
            })
        }
        res.send(result)
    })
})

app.listen(process.env.PORT || 3000, () => {
    console.log("Listening on port " + (process.env.PORT || 3000))
})