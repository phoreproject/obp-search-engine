'use strict';
const fs = require('fs'),
    path = require('path'),
    Sequelize = require('sequelize'),
    lodash = require('lodash'),
    db = {};

// const sequelize = new Sequelize(process.env.DATABASE_URI || 'mysql://' + process.env.RDS_USERNAME + ':' + process.env.RDS_PASSWORD + '@' + process.env.RDS_HOSTNAME + ':' + process.env.RDS_PORT + '/' + process.env.RDS_DB_NAME, {
//     omitNull: true
// });
const sequelize = new Sequelize('mysql://user:secret@127.0.0.1:3306/obpsearch', {omitNull: true});

fs.readdirSync(__dirname).filter(function (file) {
    return ((file.indexOf('.') !== 0) && (file !== 'index.js') && (file.slice(-3) === '.js'));
}).forEach(function (file) {
    const model = sequelize.import(path.join(__dirname, file));
    db[model.name] = model;
});

Object.keys(db).forEach(function (modelName) {
    if (db[modelName].options.hasOwnProperty('associate')) {
        db[modelName].options.associate(db);
    }
});

module.exports = lodash.extend({
    sequelize: sequelize,
    Sequelize: Sequelize
}, db);
