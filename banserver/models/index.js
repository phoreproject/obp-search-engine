'use strict';
const fs = require('fs'),
    path = require('path'),
    { Sequelize, Op } = require('sequelize'),
    lodash = require('lodash'),
    db = {};

const sequelize = new Sequelize(process.env.DATABASE_URI || 'mysql://' + process.env.RDS_USERNAME + ':' + process.env.RDS_PASSWORD + '@' + process.env.RDS_HOSTNAME + ':' + process.env.RDS_PORT + '/' + process.env.RDS_DB_NAME, {
    omitNull: true
});

fs.readdirSync(__dirname).filter(function (file) {
    return ((file.indexOf('.') !== 0) && (file !== 'index.js') && (file.slice(-3) === '.js'));
}).forEach(function (file) {
    const model_instance = require(path.join(__dirname, file))(sequelize, Sequelize);
    db[model_instance.name] = model_instance;
});

Object.keys(db).forEach(function (modelName) {
    if (db[modelName].options.hasOwnProperty('associate')) {
        db[modelName].options.associate(db);
    }
});

module.exports = lodash.extend({
    sequelize: sequelize,
    Sequelize: Sequelize,
    sequelize_Op: Op,
}, db);
