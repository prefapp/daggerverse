import { createAppAuth } from "@octokit/auth-app";
import * as fs from 'fs';
          
const auth = createAppAuth({
    appId: process.env.GITHUB_APP_ID,
    privateKey: process.env.GITHUB_APP_PEM_FILE,
    installationId: process.env.GITHUB_APP_INSTALLATION_ID,
});

const { token } = await auth({ type: 'installation' });

fs.writeFileSync("/token", token);

