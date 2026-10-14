FROM node:20-alpine

WORKDIR /app

COPY package.json pnpm-lock.yaml pnpm-workspace.yaml* ./
RUN npm install -g pnpm && pnpm install || npm install

COPY . .

EXPOSE 3000
CMD ["pnpm", "dev", "-p", "3000"]
