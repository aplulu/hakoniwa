import { useState } from 'react';
import { Card, Flex, Text } from '@radix-ui/themes';
import { PlusCircle } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface CreateInstanceCardProps {
  delayIndex: number;
  onClick: () => void;
}

export function CreateInstanceCard({ delayIndex, onClick }: CreateInstanceCardProps) {
  const { t } = useTranslation();
  const [isHovered, setIsHovered] = useState(false);

  return (
    <Card 
      size="3" 
      className="create-card anim-entry"
      style={{ 
        display: 'flex', 
        alignItems: 'center', 
        justifyContent: 'center', 
        cursor: 'pointer',
        borderStyle: 'dashed',
        borderWidth: '2px',
        borderColor: isHovered ? 'var(--accent-8)' : 'var(--gray-6)',
        backgroundColor: isHovered ? 'var(--gray-2)' : 'transparent',
        minHeight: '220px', // Aligning with other cards roughly
        animationDelay: `${delayIndex * 0.05}s`,
        transition: 'all 0.2s ease'
      }}
      onClick={onClick}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <Flex direction="column" align="center" gap="3">
        <PlusCircle 
          size={40} 
          strokeWidth={1.5} 
          color={isHovered ? 'var(--accent-9)' : 'var(--gray-8)'}
          style={{ transition: 'color 0.2s' }}
        />
        <Text 
          size="2" 
          weight="bold" 
          style={{ 
            color: isHovered ? 'var(--accent-9)' : 'var(--gray-11)',
            transition: 'color 0.2s' 
          }}
        >
          {t('workspace.create.card_title')}
        </Text>
      </Flex>
    </Card>
  );
}
