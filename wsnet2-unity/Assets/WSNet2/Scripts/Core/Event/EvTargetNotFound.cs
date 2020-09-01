using System;

namespace WSNet2.Core
{
    public class EvTargetNotFound : Event, EvMsgError
    {
        public string[] Targets { get; private set; }
        public MsgType MsgType { get; private set; }
        public int MsgSeqNum { get; private set; }
        public ArraySegment<byte> Payload { get; private set; }

        public EvTargetNotFound(SerialReader reader) : base(EvType.TargetNotFound, reader)
        {
            Targets = reader.ReadStrings();
            MsgType = (MsgType)reader.Get8();
            MsgSeqNum = reader.Get24();
            Payload = reader.GetRest();
        }
    }
}
