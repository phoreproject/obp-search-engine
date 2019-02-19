module.exports = function (sequelize, DataTypes) {
    return sequelize.define('moderatorIdsPerItem', {
            peerID: {
                type: DataTypes.STRING(50),
                allowNull: false,
                unique: true,
                primaryKey: true
            },
            moderatorID: {
                type: DataTypes.STRING(50),
                allowNull: false,
                unique: false,
                primaryKey: true
            }
        },
        {
            freezeTableName: true,
            timestamps: false
        });
};
