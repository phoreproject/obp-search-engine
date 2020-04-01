'use strict';

const express = require('express');
const app = express();
const path = require('path');
const moment = require('moment');

const ConfigCreator = require('./configCreator').ConfigCreator;
let TagCache = require('./tagsCache').TagsCache;
TagCache = new TagCache();

const ORM = require('./ORM.js');

app.get('/logo.png', (req, res) => {
    res.sendFile('logo.png', {root: path.join(__dirname)});
});

const config = require('./config');
app.get('/', (req, res) => {
    res.send(config);
});

app.get('/search/listings', async (req, res) => {
    try {
        const itemQueryOptions = {};
        const page = req.query.p || 0;
        const ps = Math.min(req.query.ps || 20, 100);
        const queryNSFW = (req.query.nsfw === 'true') || false;
        const queryOrderBy = req.query.sortBy || 'RELEVANCE';
        const queryRating = Number(req.query.rating || "0");
        const queryModerators = req.query.b2_moderators;
        const queryContractType = req.query.type;
        const testnet = (req.query.network === 'testnet') || false;

        itemQueryOptions.limit = ps;
        itemQueryOptions.offset = ps * page;
        itemQueryOptions.where = {
            blocked: false,
            testnet: testnet,
        };
        if (queryNSFW === false) { // return no nsfw results or all results
            itemQueryOptions.where.nsfw = queryNSFW;
        }

        // create query to filter by rating
        if (queryRating !== 0) {
            itemQueryOptions.where.averageRating = {
                [ORM.sequelize_Op.gte]: {
                    5: 4.75,
                    4: 4,
                    3: 3,
                    2: 2,
                    1: 0
                }[queryRating]
            };
        }

        itemQueryOptions.order = [[]];
        // create query to order by
        if (queryOrderBy.startsWith('PRICE')) {
            itemQueryOptions.order[0][0] = 'priceAmount';
        } else if (queryOrderBy.startsWith('RATING')) {
            itemQueryOptions.order[0][0] = 'averageRating';
        } else if (queryOrderBy.startsWith('NAME')) {
            itemQueryOptions.order[0][0] = 'title';
        }

        if (queryOrderBy.endsWith('DESC')) {
            itemQueryOptions.order[0][1] = 'DESC';
        } else if (queryOrderBy.endsWith('ASC')) {
            itemQueryOptions.order[0][1] = 'ASC';
        }
        if (itemQueryOptions.order[0].length === 0) {
            itemQueryOptions.order = undefined;
        }
        if (queryOrderBy === 'RAND') {
            itemQueryOptions.order = ORM.sequelize.random()
        }

        // create query to filter by searching name or tag
        if (req.query.q && req.query.q !== '*') {
            // const words = req.query.q.replace(/[^\w]/g, '').split(' ') old version, why this replace pattern?
            const words = req.query.q.split(' ').map((word) => {
                return {
                    [ORM.sequelize_Op.like]: '%' + word + '%'
                };
            });
            const oneOfWordsInTitle = {
                [ORM.sequelize_Op.or]: words
            };

            itemQueryOptions.where = {
                [ORM.sequelize_Op.or]: {
                    title: oneOfWordsInTitle,
                    tags: oneOfWordsInTitle
                }
            };
        }

        let nodeQueryWhere = {
            lastUpdated: {
                [ORM.sequelize_Op.gt]: moment(new Date()).subtract(8, 'hours').toDate()
            },
            listed: true,
            blocked: false
        };

        // create query to filter nodes by moderator / verified moderator
        if (queryModerators !== undefined) {
            if (queryModerators === 'verified_mods') {
                nodeQueryWhere.verifiedModerator = true;
            }
            else if (queryModerators === 'all_mods') {
                nodeQueryWhere.moderator = true;
            }
            // else get all
        }

        itemQueryOptions.include = [{
            model: ORM.Node,
            where: nodeQueryWhere
        }];

        if (queryContractType !== undefined && queryContractType !== 'all') {
            itemQueryOptions.where.contractType = queryContractType
        }

        // remove duplicated peerID's
        const itemQueryOutput = await ORM.Item.findAndCountAll(itemQueryOptions);
        let peerIDs = new Set();
        itemQueryOutput.rows.forEach((item) => {
            peerIDs.add(item.peerID);
        });
        peerIDs = Array.from(peerIDs);

        // search for moderators for each peerID
        let moderators = {};
        for (let i in peerIDs) {
            const moderatorQueryOptions = {
                where:
                    {
                        peerID: {
                            [ORM.sequelize_Op.eq]: peerIDs[i]
                        }
                    }
            };

            let mods = await ORM.ModeratorIdsPerItem.findAll(moderatorQueryOptions);
            if (mods !== undefined && mods.length > 0) {
                moderators[peerIDs[i]] = [];
                for (let j in mods) {
                    moderators[peerIDs[i]].push(mods[j].dataValues.moderatorID);
                }
            }
        }

        const configuration = new ConfigCreator({
            selfLink: '/search/listings',
            nsfwVisible: queryNSFW,
            itemRating: queryRating,
            queryModerators: queryModerators,
            sortBy: queryOrderBy,
            orderType: queryContractType,
            condition: undefined, // TODO this information is not available in current db schema and crawler
            shippingInfo: undefined // TODO this information is not available in current db schema and crawler
        });

        // create result dictionary
        const result = Object.assign(configuration.toJSON(), {
            results: {
                total: itemQueryOutput.count,
                morePages: itemQueryOutput.count > ps,
                results: []
            }
        });


        const safeSplit = function (str) {
            if (str === undefined) {
                return [];
            }
            return str.split(',');
        };

        for (const r of itemQueryOutput.rows) {
            let thumbnails = safeSplit(r.thumbnail);
            result.results.results.push({
                type: 'listing',
                relationships: {
                    vendor: {
                        data: {
                            userAgent: r.node.userAgent,
                            lastSeen: r.node.lastUpdated,
                            blocked: r.node.blocked,
                            peerID: r.peerID,
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
                                averageRating: parseFloat(r.node.averageRating),
                            }
                        }
                    },
                    moderators: (r.peerID in moderators) ? moderators[r.peerID] : []
                },
                data: {
                    score: r.score,
                    hash: r.hash,
                    slug: r.slug,
                    title: r.title,
                    tags: safeSplit(r.tags),
                    categories: safeSplit(r.categories),
                    contractType: r.contractType,
                    format: r.format,
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
                    averageRating: parseFloat(r.averageRating),
                    ratingCount: r.ratingCount,
                    acceptedCurrencies: safeSplit(r.acceptedCurrencies),
                    coinType: r.coinType,
                    coinDivisibility: r.coinDivisibility,
                    normalizedPrice: r.normalizedPrice
                }
            });
        }
        res.send(result);
    }
    catch (err) {
        return res.status(500).send(err);
    }
});

app.get('/search/toptags', async (req, res) => {
    try {
        res.send(TagCache.getTags(req.query.tags));
    }
    catch (err) {
        return res.status(500).send(err);
    }
});

app.get('/verified_moderators', async (req, res) => {
    const options = {};
    options.where = {
        isVerified: true
    };

    const out = await ORM.Moderators.findAll(options);
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

app.get('/healthCheck', async (req, res) => {
    res.send('');
});

const port = process.env.PORT || 3000;
app.listen(port, () => {
    console.log('Listening on port ' + (port));
});
