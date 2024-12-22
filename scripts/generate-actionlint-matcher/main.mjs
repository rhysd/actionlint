import { promises as fs } from 'node:fs';
import object from './object.mjs';

async function main(args) {
    const json = JSON.stringify(object, null, 2);
    if (args.length === 0) {
        console.log(json);
    } else {
        const path = args[0];
        await fs.writeFile(args[0], json + '\n', 'utf8');
        console.log(`Wrote to ${path}`);
    }
}

main(process.argv.slice(2));
