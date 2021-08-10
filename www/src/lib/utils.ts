export function removeFromArray<T>(arr: Array<T>, val: T): Array<T> {
	const i = arr.indexOf(val);
	if (i > -1) {
		arr.splice(i, 1);
	}
	return arr;
}
