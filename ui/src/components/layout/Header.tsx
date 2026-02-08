import {
  Box,
  Container,
  Flex,
  Heading,
  Avatar,
  Text,
  IconButton,
} from '@radix-ui/themes';
import { LogOut } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { Configuration, User } from '../../types';

interface HeaderProps {
  config?: Configuration;
  user: User;
  onLogout: () => void;
}

export function Header({ config, user, onLogout }: HeaderProps) {
  const { t } = useTranslation();

  return (
    <Box
      style={{
        position: 'sticky',
        top: 0,
        zIndex: 10,
        backgroundColor: 'rgba(255, 255, 255, 0.7)',
        backdropFilter: 'blur(12px)',
        borderBottom: '1px solid var(--gray-4)',
      }}
      px="4"
      py="3">
      <Container size="4">
        <Flex justify="between" align="center">
          <Flex align="center" gap="3">
            <img
              src={config?.logo_url || '/_hakoniwa/hakoniwa_logo.webp'}
              alt="Logo"
              style={{ width: 28, height: 28, objectFit: 'contain' }}
            />
            <Heading
              size="4"
              weight="bold"
              style={{ letterSpacing: '-0.02em' }}>
              {config?.title || 'Hakoniwa'}
            </Heading>
          </Flex>
          <Flex align="center" gap="4">
            <Flex
              align="center"
              gap="3"
              style={{
                padding: '4px 12px',
                background: 'var(--gray-3)',
                borderRadius: '99px',
              }}>
              <Avatar
                size="1"
                radius="full"
                fallback={user.id.substring(0, 2).toUpperCase()}
                color="indigo"
                variant="solid"
              />
              <Text size="2" weight="medium" color="gray">
                {user.type === 'anonymous' ? t('user.guest') : user.id}
              </Text>
            </Flex>
            <IconButton
              variant="ghost"
              color="gray"
              onClick={onLogout}
              title={t('user.logout')}>
              <LogOut size={18} />
            </IconButton>
          </Flex>
        </Flex>
      </Container>
    </Box>
  );
}
