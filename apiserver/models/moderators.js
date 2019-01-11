module.exports = function (sequelize, DataTypes) {
    return sequelize.define('moderators', {
            id: {
                type: DataTypes.STRING(50),
                allowNull: false,
                unique: true,
                primaryKey: true
            },
            type: DataTypes.STRING(16),
            isVerified: DataTypes.TINYINT(1),
        },
        {
            freezeTableName: true,
            timestamps: false
        });
};
