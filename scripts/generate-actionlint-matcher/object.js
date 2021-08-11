const E = '(?:\\x1b\\[\\d+m)'; // Matching to ANSI color escape sequence
const FILEPATH = '(.+)';
const LINE = '(\\d+)';
const COL = '(\\d+)';
const MESSAGE = '(.+)';
const KIND = '\\[(.+)\\]';

const regexp = `^${E}?${FILEPATH}${E}*:${E}*${LINE}${E}*:${E}*${COL}${E}*: ${E}*${MESSAGE}${E}*${KIND}$`;
const object = {
    problemMatcher: [
        {
            owner: 'actionlint',
            pattern: [
                {
                    regexp,
                    file: 1,
                    line: 2,
                    column: 3,
                    message: 4,
                    code: 5,
                },
            ],
        },
    ],
};

module.exports = object;
