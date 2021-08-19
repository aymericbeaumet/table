import Entity from './entity';

test('entity inference', () => {
	const tests = [
		// Text
		{
			in: 'lorem ipsum',
			expected: [],
		},

		// SoundCloud
		{
			in: 'Listen to this https://soundcloud.com/darealartis/exclu-tiitof-bad-boy!',
			expected: [
				new Entity('iframe', {
					icon: 'soundcloud',
					src: 'https://w.soundcloud.com/player/?url=https%3A%2F%2Fsoundcloud.com%2Fdarealartis%2Fexclu-tiitof-bad-boy&auto_play=false&show_artwork=true',
				}),
			],
		},
	];

	for (const t of tests) {
		// A text entity is always the last option
		t.expected.push(new Entity('text', { value: t.in }));

		const out = Entity.infer(t.in);
		expect(out).toEqual(t.expected);
	}
});
