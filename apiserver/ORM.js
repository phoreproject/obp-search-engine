'use strict';

const { Sequelize, Op } = require('sequelize');
const sequelize = new Sequelize(process.env.DATABASE_URI || 'mysql://' + process.env.RDS_USERNAME + ':' + process.env.RDS_PASSWORD + '@' + process.env.RDS_HOSTNAME + ':' + process.env.RDS_PORT + '/' + process.env.RDS_DB_NAME, {omitNull: true});

const Item = require('./models/item')(sequelize, Sequelize);
const Node = require('./models/node')(sequelize, Sequelize);
const Moderators = require('./models/moderators')(sequelize, Sequelize);
const ModeratorIdsPerItem = require('./models/moderatorIdsPerItem')(sequelize, Sequelize);

Item.belongsTo(Node, {foreignKey: 'peerID'}); // just to hide warning.

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
