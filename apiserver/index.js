'use strict';

const express = require('express');
const app = express();
const Sequelize = require('sequelize');

const path = require('path');
const moment = require('moment');

// const sequelize = new Sequelize(process.env.DATABASE_URI || 'mysql://' + process.env.RDS_USERNAME + ':' + process.env.RDS_PASSWORD + '@' + process.env.RDS_HOSTNAME + ':' + process.env.RDS_PORT + '/' + process.env.RDS_DB_NAME, {omitNull: true});
const sequelize = new Sequelize('mysql://user:secret@127.0.0.1:3306/obpsearch', {omitNull: true});

const Item = sequelize.import('./models/item');
const Node = sequelize.import('./models/node');
const Moderator = sequelize.import('./models/moderator');

Item.belongsTo(Node, {foreignKey: 'owner'});

app.get('/logo.png', (req, res) => {
    res.sendFile('logo.png', {root: path.join(__dirname)});
});

const config = require('./config');

app.get('/', (req, res) => {
    res.send(config);
});

app.get('/search/listings', (req, res) => {
    const options = {};
    const page = req.query.p || 0;
    const ps = Math.min(req.query.ps || 20, 100);
    const nsfw = req.query.nsfw || false;
    const orderBy = req.query.sortBy || 'RELEVANCE';

    options.limit = ps;
    options.offset = ps * page;
    options.where = {};
    options.where.nsfw = nsfw;

    if (req.query.rating) {
        options.where.rating = {
            [sequelize.Op.gte]: {
                5: 4.75,
                4: 4,
                3: 3,
                2: 2,
                1: 0
            }[Number(req.query.rating)]
        };
    }
    options.order = [[]];

    if (orderBy.startsWith('PRICE')) {
        options.order[0][0] = 'priceAmount';
    } else if (orderBy.startsWith('RATING')) {
        options.order[0][0] = 'rating';
    }
    if (orderBy.endsWith('DESC')) {
        options.order[0][1] = 'DESC';
    } else if (orderBy.endsWith('ASC')) {
        options.order[0][1] = 'ASC';
    }
    if (options.order[0].length === 0) {
        options.order = undefined;
    }
    console.log(req.query.q);

    if (req.query.q && req.query.q !== '*') {
        // const words = req.query.q.replace(/[^\w]/g, '').split(' ') old version, why this replace pattern?
        const words = req.query.q.split(' ').map((word) => {
            return {
                [sequelize.Op.like]: '%' + word + '%'
            };
        });
        const oneOfWordsInTitle = {
            [sequelize.Op.or]: words
        };

        options.where = {
            [sequelize.Op.or]: {
                title: oneOfWordsInTitle,
                tags: oneOfWordsInTitle
            }
        };
    }

    options.include = [{
        model: Node,
        where: {
            lastUpdated: {
                [sequelize.Op.gt]: moment(new Date()).subtract(8, 'hours').toDate()
            },
            listed: true,
            blocked: false
        }
    }];

    Item.findAndCountAll(options).then((out) => {
        const result = Object.assign(config, {
            results: {
                total: out.count,
                morePages: out.count > ps,
                results: []
            }
        });
        for (const r of out.rows) {
            let thumbnails = r.thumbnail.split(',');
            result.results.results.push({
                type: 'listing',
                relationships: {
                    vendor: {
                        data: {
                            userAgent: r.node.userAgent,
                            lastSeen: r.node.lastUpdated,
                            blocked: r.node.blocked,
                            peerID: r.owner,
                            name: r.node.name,
                            handle: r.node.handle,
                            location: r.node.location,
                            nsfw: r.node.nsfw,
                            vendor: r.node.vendor,
                            moderator: r.node.moderator,
                            verifiedModerator: r.node.verifiedModerator,
                            about: r.node.about,
                            shortDescription: r.node.shortDescription,
                            avatarHashes: {
                                tiny: r.node.avatarTinyHash,
                                small: r.node.avatarSmallHash,
                                medium: r.node.avatarMediumHash,
                                original: r.node.avatarOriginalHash,
                                large: r.node.avatarLargeHash
                            },
                            headerHashes: {
                                tiny: r.node.headerTinyHash,
                                small: r.node.headerSmallHash,
                                medium: r.node.headerMediumHash,
                                original: r.node.headerOriginalHash,
                                large: r.node.headerLargeHash
                            },
                            stats: {
                                followerCount: r.node.followerCount,
                                followingCount: r.node.followingCount,
                                listingCount: r.node.listingCount,
                                postCount: r.node.postCount,
                                ratingCount: r.node.ratingCount,
                                averageRating: r.node.averageRating,
                            }
                        }
                    },
                    moderators: []
                },
                data: {
                    score: r.score,
                    hash: r.hash,
                    slug: r.slug,
                    title: r.title,
                    tags: r.tags.split(','),
                    categories: r.categories.split(','),
                    contractType: r.contractType,
                    description: r.description,
                    thumbnail: {
                        tiny: thumbnails[0],
                        small: thumbnails[1],
                        medium: thumbnails[2],
                        original: thumbnails[3],
                        large: thumbnails[4]
                    },
                    language: r.language,
                    price: {
                        amount: r.priceAmount,
                        currencyCode: r.priceCurrency,
                        modifier: r.priceModifier
                    },
                    nsfw: r.nsfw,
                    averageRating: r.averageRating,
                    ratingCount: r.ratingCount,
                    coinType: r.coinType,
                    coinDivisibility: r.coinDivisibility,
                    normalizedPrice: r.normalizedPrice
                }
            });
        }
        res.send(result);
    });
});

app.get('/verified_moderators', (req, res) => {
    const options = {};
    options.where = {
        isVerified: true
    };
    Moderator.findAll(options).then((out) => {
        const result = {
            data: {
                name: 'Marketplace',
                description: '',
                link: 'https://search.phore.io/verified_moderators.html'
            },
            types: [
                {
                    name: 'standard',
                    description: 'A moderator that has been vetted by Phore',
                    badge: {
                        tiny: 'https://search.ob1.io/images/verified_moderator_badge_tiny.png',
                        small: 'https://search.ob1.io/images/verified_moderator_badge_small.png',
                        medium: 'https://search.ob1.io/images/verified_moderator_badge_medium.png',
                        large: 'https://search.ob1.io/images/verified_moderator_badge_large.png',
                    }
                }
            ],
            moderators: [],
        };
        if (out.length !== 0) {
            for (const r of out) {
                result.moderators.push({
                    peerID: r.id,
                    type: r.type,
                });
            }
        }
        res.send(result);
    });
});


app.listen(process.env.PORT || 3000, () => {
    console.log('Listening on port ' + (process.env.PORT || 3000));
});
