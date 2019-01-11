module.exports = function (sequelize, DataTypes) {
    return sequelize.define('moderatorIdsPerItem', {
            peerID: {
                type: DataTypes.STRING(50),
                allowNull: false,
                unique: true,
                primaryKey: true
            },
            itemDataBaseID: {
                type: DataTypes.INTEGER,
                allowNull: false,
                unique: true,
                primaryKey: true
            },
            moderatorID: {
                type: DataTypes.STRING(50),
                allowNull: false,
                unique: true,
                primaryKey: true
            }
        },
        {
            freezeTableName: true,
            timestamps: false
        });
};
