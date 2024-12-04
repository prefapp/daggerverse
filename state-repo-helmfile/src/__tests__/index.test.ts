import { StateRepoHelmfile } from '../index';
import { describe, it } from 'node:test'
describe('StateRepoHelmfile', () => {


	it('hello world!', () => {
		expect(1 + 1).toBe(2);
	});
	it('containerEcho should return a container that echoes the provided string', async () => {
		const stateRepoHelmfile = new StateRepoHelmfile();
		const container = stateRepoHelmfile.containerEcho("hello world");
		const result = await container.stdout();
		expect(result).toBe("hello world\n");
	});
});
