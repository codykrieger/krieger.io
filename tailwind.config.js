module.exports = {
    future: {
        removeDeprecatedGapUtilities: true,
        purgeLayersByDefault: true,
    },
    purge: {
        // enabled: true,
        content: [
            './**/*.html',
        ],
    },
    theme: {
        extend: {},
        fontFamily: {
            // display: ['Gilroy', 'sans-serif'],
            // body: ['Graphik', 'sans-serif'],
        },
    },
    variants: {},
    plugins: [
        require('@tailwindcss/typography'),
    ],
}
