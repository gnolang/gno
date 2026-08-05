// Identicon for a bech32 address: a port of ethereum-blockies-base64's
// algorithm, rendered as inline SVG rects instead of a base64 PNG. It scales,
// needs no encoder, and is less code than the original.

const SIZE = 8;

// 32-bit xorshift seeded from the full address. Deterministic per address, so
// the same account always draws the same avatar.
function makeRand(address: string): () => number {
	const seed = [0, 0, 0, 0];
	for (let i = 0; i < address.length; i++) {
		seed[i % 4] =
			((seed[i % 4] << 5) - seed[i % 4] + address.charCodeAt(i)) | 0;
	}
	return () => {
		const t = seed[0] ^ (seed[0] << 11);
		seed[0] = seed[1];
		seed[1] = seed[2];
		seed[2] = seed[3];
		seed[3] = (seed[3] ^ (seed[3] >> 19) ^ t ^ (t >> 8)) | 0;
		return (seed[3] >>> 0) / ((1 << 31) >>> 0);
	};
}

function makeColor(rand: () => number): string {
	const h = Math.floor(rand() * 360);
	const s = rand() * 60 + 40;
	const l = (rand() + rand() + rand() + rand()) * 25;
	return `hsl(${h} ${s.toFixed(1)}% ${l.toFixed(1)}%)`;
}

// avatarSpec is the pure half: the three colours and the 8x8 grid, with no DOM
// involved, so it can be exercised outside a browser.
export function avatarSpec(address: string): {
	background: string;
	cells: string[];
} {
	const rand = makeRand(address);
	// Draw order matters — it is what makes this reproduce blockies.
	const primary = makeColor(rand);
	const background = makeColor(rand);
	const spot = makeColor(rand);

	const half = Math.ceil(SIZE / 2);
	const cells: string[] = [];
	for (let y = 0; y < SIZE; y++) {
		const row: number[] = [];
		for (let x = 0; x < half; x++) {
			row.push(Math.floor(rand() * 2.3));
		}
		// Mirrored horizontally, which is what gives a blockie its face-like shape.
		const full = row.concat(row.slice(0, SIZE - half).reverse());
		for (const value of full) {
			cells.push(value === 0 ? background : value === 1 ? primary : spot);
		}
	}
	return { background, cells };
}

// avatarSVG returns a self-contained SVG. Every interpolated value comes from
// the numeric PRNG above, never from the address text, so there is no path
// from the address into markup.
export function avatarSVG(address: string): string {
	const { background, cells } = avatarSpec(address);
	const rects = cells
		.map((color, i) =>
			color === background
				? ""
				: `<rect x="${i % SIZE}" y="${Math.floor(i / SIZE)}" width="1" height="1" fill="${color}"/>`,
		)
		.join("");
	return `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${SIZE} ${SIZE}" shape-rendering="crispEdges" role="img" aria-hidden="true"><rect width="${SIZE}" height="${SIZE}" fill="${background}"/>${rects}</svg>`;
}
