'use strict';

const moment = require('moment');

const ORM = require('./ORM.js');

const MaxDefaultTags = 11;
const UpdateIntervalDefault = 12 * 60 * 60 * 1000;
const BatchSize = 100;

class TagsCache {
    constructor(maxTags, updateInterval, updateBatchSize) {
        this._defaultTags = {
            tags: [
                'Art',
                'Music',
                'Toys',
                'Crypto',
                'Books',
                'Health',
                'Games',
                'Handmade',
                'Clothing',
                'Electronics',
                'Phore',
            ],
            count: {
                'Art': undefined,
                'Music': undefined,
                'Toys': undefined,
                'Crypto': undefined,
                'Books': undefined,
                'Health': undefined,
                'Games': undefined,
                'Handmade': undefined,
                'Clothing': undefined,
                'Electronics': undefined,
                'Phore': undefined,
            }
        };

        this.tags = undefined;
        this.updateInterval = updateInterval || UpdateIntervalDefault;
        this.batchSize = updateBatchSize || BatchSize;
        this.maxTags = maxTags || MaxDefaultTags;

        this.updateTags();
        setInterval(this.updateTags, this.updateInterval);
    }

    async updateTags() {
        const batchSize = this.batchSize;

        const itemQueryOptions = {};
        itemQueryOptions.limit = batchSize;
        itemQueryOptions.offset = 0;
        itemQueryOptions.where = {};

        let nodeQueryWhere = {
            lastUpdated: {
                [ORM.sequelize.Op.gt]: moment(new Date()).subtract(24, 'hours').toDate()
            },
            listed: true,
            blocked: false
        };

        itemQueryOptions.include = [{
            model: ORM.Node,
            where: nodeQueryWhere,
        }];

        let localTags = {};
        for (let page = 1; ; page++) {
            const itemQueryOutput = await ORM.Item.findAll(itemQueryOptions);
            if (itemQueryOutput === undefined || itemQueryOutput.length === 0) {
                break
            }
            itemQueryOptions.offset = batchSize * page;
            itemQueryOptions.limit = batchSize * (page + 1);

            for (const r of itemQueryOutput) {
                const t = r.tags.split(',');
                for (const tag of t) {
                    if (tag === '') {
                        continue
                    }
                    const capitalizeTag = tag.charAt(0).toUpperCase() + tag.substring(1).toLowerCase();
                    if (localTags.hasOwnProperty(capitalizeTag)) {
                        localTags[capitalizeTag]++;
                    } else {
                        localTags[capitalizeTag] = 1;
                    }
                }
            }
        }
        this.tags = localTags;
    }

    getTags(maxTags) {
        const max = maxTags || this.maxTags;

        if (this.tags === undefined) {
            return JSON.stringify(this._defaultTags);
        } else {
            let tags = this.tags;
            tags = Object.keys(tags).map(function (key) {
                return [key, tags[key]];
            });
            tags.sort(function (first, second) {
                return second[1] - first[1];
            });

            let jsonTags = {
                tags: [],
                count: {},
            };

            for (const t of tags.slice(0, max)) {
                jsonTags.tags.push(t[0]);
                jsonTags.count[t[0]] = t[1];
            }

            return JSON.stringify(jsonTags);
        }
    }
}

module.exports = {
    TagsCache: TagsCache,
};