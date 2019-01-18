'use strict';

class ConfigCreator {
    constructor(selfLink, nsfwVisible, itemRating, queryModerators, sortBy, orderType) {
        this.PHORE_WEBSITE = 'localhost';
        if (nsfwVisible === undefined && typeof selfLink === 'object') {
            const dictOfValues = selfLink;
            this.selfLink = dictOfValues['selfLink'];
            this.nsfwVisible = dictOfValues['nsfwVisible'];
            this.itemRating = dictOfValues['itemRating'];
            this.queryModerators = dictOfValues['queryModerators'];
            this.sortBy = dictOfValues['sortBy'];
            this.orderType = dictOfValues['orderType'];
        }
        else {
            this.selfLink = selfLink;
            this.nsfwVisible = nsfwVisible;
            this.itemRating = itemRating;
            this.queryModerators = queryModerators;
            this.sortBy = sortBy;
            this.orderType = orderType;
        }
    }

    toJSON() {
        return {
            'name': 'Phore Search',
            'logo': this.PHORE_WEBSITE + 'logo.png',

            'links': {
                'self': this.selfLink,
                'listings': this.PHORE_WEBSITE + '/search/listings'
            },
            'options': {
                'nsfw':{
                    'type':'radio',
                    'label':'Adult Content',
                    'options':[
                        {
                            'value':'true',
                            'label':'Visible',
                            'checked': this.nsfwVisible === true,
                            'default':false
                        },
                        {
                            'value':'false',
                            'label':'Hidden',
                            'checked': this.nsfwVisible !== true,
                            'default':true
                        }
                    ]
                },
                'rating': {
                    'type': 'radio',
                    'label': 'Rating',
                    'options': [
                        {
                            'value': '0',
                            'label': 'All',
                            'checked': this.itemRating === 0 || this.itemRating === undefined,
                            'default': true
                        },
                        {
                            'value': '5',
                            'label': '⭐⭐⭐⭐⭐',
                            'checked': this.itemRating === 5,
                            'default': false
                        },
                        {
                            'value': '4',
                            'label': '⭐⭐⭐⭐ \u0026 up',
                            'checked': this.itemRating === 4,
                            'default': false
                        },
                        {
                            'value': '3',
                            'label': '⭐⭐⭐ \u0026 up',
                            'checked': this.itemRating === 3,
                            'default': false
                        },
                        {
                            'value': '2',
                            'label': '⭐⭐ \u0026 up',
                            'checked': this.itemRating === 2,
                            'default': false
                        },
                        {
                            'value': '1',
                            'label': '⭐ \u0026 up',
                            'checked': this.itemRating === 1,
                            'default': false
                        }
                    ]
                },
                "b2_moderators":{
                    "type":"radio",
                    "label":"Moderation",
                    "options":[
                        {
                            "value":"verified_mods",
                            "label":"Phore Verified Moderators",
                            "checked": this.queryModerators === 'verified_mods',
                            "default":false
                        },
                        {
                            "value":"all_mods",
                            "label":"All Moderators",
                            "checked": this.queryModerators === 'all_mods',
                            "default":false
                        },
                        {
                            "value":"all_listings",
                            "label":"All Listings",
                            "checked": this.queryModerators === undefined || this.queryModerators === 'all_listings',
                            "default":true
                        }
                    ]
                },
                "type":{
                    "type":"radio",
                    "label":"Type",
                    "options":[
                        {
                            "value":"ANY",
                            "label":"Any",
                            "checked":this.orderType === 'ANY' || this.orderType === undefined,
                            "default":true
                        },
                        {
                            "value":"PHYSICAL_GOOD",
                            "label":"Physical Goods",
                            "checked":this.orderType === 'PHYSICAL_GOOD',
                            "default":false
                        },
                        {
                            "value":"CRYPTOCURRENCY",
                            "label":"Cryptocurrency",
                            "checked":this.orderType === 'CRYPTOCURRENCY',
                            "default":false
                        },
                        {
                            "value":"DIGITAL_GOOD",
                            "label":"Digital Goods",
                            "checked":this.orderType === 'DIGITAL_GOOD',
                            "default":false
                        },
                        {
                            "value":"SERVICE",
                            "label":"Services",
                            "checked":this.orderType === 'SERVICE',
                            "default":false
                        }
                    ]
                },
            },
            'sortBy': {
                'PRICE_ASC': {
                    'label': 'Price (Low to High)',
                    'selected': this.sortBy === 'PRICE_ASC',
                    'default': false
                },
                'PRICE_DESC': {
                    'label': 'Price (High to Low)',
                    'selected': this.sortBy === 'PRICE_DESC',
                    'default': false
                },
                'NAME_ASC': {
                    'label': 'Name (Ascending)',
                    'selected': this.sortBy === 'NAME_ASC',
                    'default': false
                },
                'NAME_DESC': {
                    'label': 'Name (Descending)',
                    'selected': this.sortBy === 'NAME_DESC',
                    'default': false
                },
                'RATING_ASC': {
                    'label': 'Rating (Low to High)',
                    'selected': this.sortBy === 'RATING_ASC"',
                    'default': false
                },
                'RATING_DESC': {
                    'label': 'Rating (High to Low)',
                    'selected': this.sortBy === 'RATING_DESC',
                    'default': false
                },
                'RELEVANCE': {
                    'label': 'Relevance',
                    'select': this.sortBy === "RELEVANCE" || this.sortBy === undefined,
                    'default': true
                }
            }
        }
    }
}

module.exports = {
    ConfigCreator: ConfigCreator
};