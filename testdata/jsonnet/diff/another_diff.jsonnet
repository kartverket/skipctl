{
  person: {
    name: 'Alice',
    age: 30 + 12 + 'bob',
    hobbies: [
      'cycling',
    ] + ['reading'],
  },
} + {
  person+: {
    hobbies+: [
      'swimming',
    ],
  },
}
