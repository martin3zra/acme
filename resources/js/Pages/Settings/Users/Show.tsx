import { User } from '@/types';

export type Props = {
  user: User;
};
export default function ShowUser({ user }: Props) {
  return <h1>Display User: {user.name}</h1>;
}
