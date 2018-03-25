module.exports = function(sequelize, DataTypes){
    return sequelize.define('items', {
        owner: DataTypes.STRING(50),
        hash: {
            type: DataTypes.STRING(50),
            allowNull: false,
            unique: false,
            primaryKey: true
        },
        slug: DataTypes.STRING(70),
        title: DataTypes.STRING(140),
        tags: DataTypes.STRING(410),
        description: DataTypes.STRING(50000),
        thumbnail: DataTypes.STRING(160),
        language: DataTypes.STRING(20),
        priceAmount: DataTypes.BIGINT,
        priceCurrency: DataTypes.STRING(10),
        categories: DataTypes.STRING(410),
        nsfw: DataTypes.BOOLEAN,
        contractType: DataTypes.STRING(20),
        rating: DataTypes.DECIMAL(3, 2)
    }, {
        freezeTableName: true,
        timestamps: false
    })
};