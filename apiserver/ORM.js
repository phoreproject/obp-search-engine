'use strict';

const Sequelize = require('sequelize');

const sequelize = new Sequelize(process.env.DATABASE_URI || 'mysql://' + process.env.RDS_USERNAME + ':' + process.env.RDS_PASSWORD + '@' + process.env.RDS_HOSTNAME + ':' + process.env.RDS_PORT + '/' + process.env.RDS_DB_NAME, {omitNull: true});
const Item = sequelize.import('./models/item');
const Node = sequelize.import('./models/node');
const Moderators = sequelize.import('./models/moderators');
const ModeratorIdsPerItem = sequelize.import('./models/moderatorIdsPerItem');

Item.belongsTo(Node, {foreignKey: 'peerID'});

module.exports = {
    sequelize: sequelize,
    Item: Item,
    Node: Node,
    Moderators: Moderators,
    ModeratorIdsPerItem: ModeratorIdsPerItem,
};