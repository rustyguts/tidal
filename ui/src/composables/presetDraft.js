// blankSpec returns a sensible default V2 spec for a brand-new preset. Saves
// the editor from a screenful of empty fields when the user clicks "New".
export function blankSpec() {
    return {
        schemaVersion: 2,
        container: { format: 'mp4', faststart: true },
        video: {
            codec: 'libx264',
            preset: 'slow',
            pixelFormat: 'yuv420p',
            rate: { mode: 'crf', crf: 23 },
            twoPass: false,
            resolution: null
        },
        audio: { codec: 'aac', bitrate: '192k' },
        filters: { video: [], audio: [] },
        subtitles: { mode: 'copy' },
        outputSuffix: '',
        rawExtras: []
    };
}
// debounce returns a function that delays calls by `wait` ms, only firing
// once after `wait` of inactivity.
export function debounce(fn, wait) {
    let t;
    return ((...args) => {
        if (t)
            clearTimeout(t);
        t = setTimeout(() => fn(...args), wait);
    });
}
