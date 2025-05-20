import React, { useEffect, useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import tribeService from '../services/tribeService';

const TribeMemberProfileScreen = () => {
  const { tribeId, memberId } = useParams();
  const [member, setMember] = useState(null);
  const [tribe, setTribe] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [inviter, setInviter] = useState(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError('');
        const tribeData = await tribeService.getTribeById(tribeId);
        setTribe(tribeData);
        const members = await tribeService.getTribeMembers(tribeId);
        const found = members.find(m => m.user_id === memberId);
        setMember(found);
        
        // If there's an inviter, find their details
        if (found && found.invited_by) {
          const inviterMember = members.find(m => m.user_id === found.invited_by);
          if (inviterMember) {
            setInviter(inviterMember);
          }
        }
      } catch (err) {
        setError('Failed to load member info');
      } finally {
        setLoading(false);
      }
    };
    fetchData();
  }, [tribeId, memberId]);

  if (loading) return <div>Loading...</div>;
  if (error || !member) return <div>{error || 'Member not found'}</div>;

  // Get inviter display name
  const inviterName = inviter ? (inviter.user?.name || inviter.display_name || 'Unknown User') : 'Unknown';

  return (
    <div className="container tribe-member-profile">
      <h2>Tribe Member Profile</h2>
      <h3>{member.user?.name || member.display_name}</h3>
      <p><strong>Email:</strong> {member.user?.email || 'Unknown'}</p>
      <p><strong>Membership type:</strong> {member.membership_type}</p>
      <p><strong>Invited by:</strong> {member.invited_by ? (
        <Link to={`/tribes/${tribeId}/members/${member.invited_by}`}>View {inviterName}'s Profile</Link>
      ) : 'Unknown'}</p>
      <p><strong>Invited at:</strong> {member.invited_at ? new Date(member.invited_at).toLocaleDateString() : 'Unknown'}</p>
      <p><Link to={`/tribes/${tribeId}`}>Back to Tribe</Link></p>
    </div>
  );
};

export default TribeMemberProfileScreen; 