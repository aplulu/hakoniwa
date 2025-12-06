import { useState } from 'react';
import { Card, Box, Flex, Heading, DropdownMenu, IconButton, Text, Button } from '@radix-ui/themes';
import { Terminal, MoreVertical, Trash2, Loader2 } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import type { Instance, InstanceType } from '../../types';

interface InstanceCardProps {
  instance: Instance;
  typeInfo?: InstanceType;
  index: number;
  onDelete: (id: string) => void;
  onOpen: (id: string) => void;
}

export function InstanceCard({ instance, typeInfo, index, onDelete, onOpen }: InstanceCardProps) {
  const { t } = useTranslation();
  const [isHovered, setIsHovered] = useState(false);

  return (
    <Card 
      size="3" 
      className="instance-card anim-entry"
      style={{ 
        position: 'relative', 
        overflow: 'visible',
        animationDelay: `${index * 0.05}s`,
        padding: '20px',
        display: 'flex',
        flexDirection: 'column',
        justifyContent: 'space-between',
        transition: 'all 0.3s ease',
        cursor: 'default',
        transform: isHovered ? 'translateY(-4px)' : 'translateY(0)',
        boxShadow: isHovered ? 'var(--shadow-4)' : 'var(--shadow-2)',
      }}
      onMouseEnter={() => setIsHovered(true)}
      onMouseLeave={() => setIsHovered(false)}
    >
      <Flex direction="column" gap="3">
        <Flex justify="between" align="start" gap="3">
           <Flex align="start" gap="3" style={{ flex: 1, minWidth: 0 }}>
             <Box 
               style={{ 
                 width: 40, 
                 height: 40, 
                 flexShrink: 0,
                 borderRadius: 'var(--radius-3)',
                 overflow: 'hidden',
                 background: 'var(--gray-3)',
                 display: 'flex',
                 alignItems: 'center',
                 justifyContent: 'center'
               }}
             >
               {typeInfo?.logo_url ? (
                 <img src={typeInfo.logo_url} alt={typeInfo.name} style={{ width: '100%', height: '100%', objectFit: 'contain' }} />
               ) : (
                 <Terminal size={20} color="var(--gray-11)" />
               )}
             </Box>
             <Box style={{ minWidth: 0 }}>
               <Heading size="4" weight="bold" style={{ whiteSpace: 'nowrap', overflow: 'hidden', textOverflow: 'ellipsis' }}>
                 {instance.name}
               </Heading>
             </Box>
           </Flex>

           <DropdownMenu.Root>
            <DropdownMenu.Trigger>
              <IconButton variant="ghost" color="gray" size="1" style={{ margin: '-4px -4px 0 0' }}>
                <MoreVertical size={16} />
              </IconButton>
            </DropdownMenu.Trigger>
            <DropdownMenu.Content>
              <DropdownMenu.Item color="red" onClick={() => onDelete(instance.id)}>
                <Trash2 size={14} style={{ marginRight: 8 }} />
                {t('workspace.action.delete')}
              </DropdownMenu.Item>
            </DropdownMenu.Content>
          </DropdownMenu.Root>
        </Flex>

        <Flex direction="column" gap="2">
          <Text size="2" color="gray">
            {typeInfo?.name || instance.type}
          </Text>
          
          <Flex align="center" gap="2">
             <Box 
               style={{ 
                 width: 8, 
                 height: 8, 
                 borderRadius: '50%', 
                 backgroundColor: instance.status === 'running' ? 'var(--green-9)' : instance.status === 'pending' ? 'var(--orange-9)' : 'var(--red-9)',
               }} 
             />
             <Text size="2" weight="medium" style={{ 
               color: instance.status === 'running' ? 'var(--green-9)' : instance.status === 'pending' ? 'var(--orange-9)' : 'var(--red-9)',
               textTransform: 'capitalize' 
             }}>
               {t(`workspace.status.${instance.status}` as any)}
             </Text>
          </Flex>
        </Flex>
      </Flex>

      <Box mt="4">
         {instance.status === 'running' ? (
           <Button 
             size="3" 
             variant="solid" 
             style={{ width: '100%', cursor: 'pointer' }}
             onClick={() => onOpen(instance.id)}
           >
             {t('workspace.action.open')}
           </Button>
         ) : (
            <Button 
              size="3" 
              disabled 
              variant="soft" 
              color="gray" 
              style={{ width: '100%' }}
            >
              {instance.status === 'pending' ? (
                <Loader2 className="animate-spin" size={16} style={{ marginRight: 8 }} />
              ) : null}
              {t(`workspace.status.${instance.status}` as any)}
            </Button>
         )}
      </Box>
    </Card>
  );
}
