'use strict';

const moment = require('moment');

const ORM = require('./ORM.js');

const MaxDefaultTags = 10;
const UpdateIntervalDefault = 12 * 60 * 60 * 1000;
const BatchSize = 100;

export class TagsCache {
    constructor(maxTags, updateInterval, updateBatchSize) {
        this.defaultTags = [
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
        ];

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
                [sequelize.Op.gt]: moment(new Date()).subtract(24, 'hours').toDate()
            },
            listed: true,
            blocked: false
        };

        itemQueryOptions.include = [{
            model: ORM.Node,
            where: nodeQueryWhere
        }];

        let localTags = {};
        for (let page = 0;;page++) {
            const itemQueryOutput = await ORM.Item.findAndCountAll(itemQueryOptions);
            if (itemQueryOptions.count === 0) {
                break
            }
            itemQueryOptions.offset = batchSize * page;

            for (const r of itemQueryOutput.rows) {
                const t = r.tags.split(',');
                for (const tag of t) {
                    const capitalizeTag = tag.charAt(0).toUpperCase() + tag.substring(1).toLowerCase();
                    if (localTags.hasOwnProperty(capitalizeTag)) {
                        localTags[capitalizeTag]++;
                    } else {
                        localTags[capitalizeTag] = 0;
                    }
                }
            }
        }
        this.tags = localTags;
    }

    getTags(maxTags) {
        const max = maxTags || this.maxTags;

        if (this.tags === undefined) {
            return JSON.stringify(this.defaultTags);
        } else {


            let tags = Object.keys(this.tags).map(function(key) {
                return [key, this.tags[key]];
            });
            tags.sort(function(first, second) {
                return second[1] - first[1];
            });

            return JSON.stringify(tag.slice(0, max));
        }
    }
}
