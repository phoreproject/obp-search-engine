'use strict';

const { Sequelize, Op } = require('sequelize');
const sequelize = new Sequelize(process.env.DATABASE_URI || 'mysql://' + process.env.RDS_USERNAME + ':' + process.env.RDS_PASSWORD + '@' + process.env.RDS_HOSTNAME + ':' + process.env.RDS_PORT + '/' + process.env.RDS_DB_NAME, {omitNull: true});
const Item = sequelize.import('./models/item');
const Node = sequelize.import('./models/node');
const Moderators = sequelize.import('./models/moderators');
const ModeratorIdsPerItem = sequelize.import('./models/moderatorIdsPerItem');


module.exports = {
    sequelize: sequelize,
    sequelize_Op: Op,
    Item: Item,
    Node: Node,
    Moderators: Moderators,
    ModeratorIdsPerItem: ModeratorIdsPerItem,
    Synced: sequelize.sync(),
};

module.exports.Synced.then(() => {
    Item.belongsTo(Node, {foreignKey: 'peerID'});
});
