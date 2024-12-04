module.exports = {
  injectGlobals: true,
  testEnvironment: 'node',
  collectCoverage: true,
  coverageDirectory: 'coverage',
  testMatch: ['**/src/**/*.test.js', '**/src/**/*.test.ts'],
  transform: {
    '^.+\\.ts$': 'ts-jest'
  }
};
