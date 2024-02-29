const ESCAPE = '(?:\\x1b\\[\\d+m)'; // Matching to ANSI color escape sequence
const FILEPATH = '.+?';
const LINE = '\\d+';
const COL = '\\d+';
const MESSAGE = '.+?';
const KIND = '.+?';

let regexp = '^E?(F)E*:E*(L)E*:E*(C)E*: E*(M)E* \\[(K)\\]$';
regexp = regexp.replaceAll('E', ESCAPE);
regexp = regexp.replace('F', FILEPATH);
regexp = regexp.replace('L', LINE);
regexp = regexp.replace('C', COL);
regexp = regexp.replace('M', MESSAGE);
regexp = regexp.replace('K', KIND);

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
