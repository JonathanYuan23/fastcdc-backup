import numpy as np

from transformers import AutoTokenizer
from sentence_transformers import SentenceTransformer

def batched(text, tokenizer, n):
    """Batch data into chunks of length n. The last batch may be shorter."""
    # batched('ABCDEFG', 3) --> ABC DEF G
    if n < 1:
        raise ValueError('n must be at least one')
    
    len_tokens = len(tokenizer.encodings[0].tokens)
    for i in range(0, len_tokens, n):
        end_token_i = i + n - 1
        if i + n >= len_tokens:
            end_token_i = len_tokens - 1

        yield (text[tokenizer.encodings[0].offsets[i][0] : tokenizer.encodings[0].offsets[end_token_i][1]], end_token_i - i + 1)

def chunked_tokens(text, encoding_name, chunk_length):
    tokenizer = AutoTokenizer.from_pretrained(encoding_name)
    tokens = tokenizer(text, add_special_tokens=False)
    chunks_iterator = batched(text, tokens, chunk_length)
    yield from chunks_iterator

def len_safe_get_embedding(text, model="thenlper/gte-small", encoding_name="thenlper/gte-small", max_tokens=512, average=True):
    model = SentenceTransformer(model)

    chunk_embeddings = []
    chunk_lens = []
    for chunk in chunked_tokens(text, encoding_name=encoding_name, chunk_length=max_tokens):
        chunk_embeddings.append(model.encode(chunk[0]))
        chunk_lens.append(chunk[1])

    if average:
        weighted_average = np.average(chunk_embeddings, axis=0, weights=chunk_lens)
        normalized_weighted_average = weighted_average / np.linalg.norm(weighted_average)  # normalizes length to 1
        # normalized_weighted_average = normalized_weighted_average.tolist()
    return normalized_weighted_average

text = '''o!
    The shirt of Nessus is upon me; teach me,
    Alcides, thou mine ancestor, thy rage;
    Let me lodge Lichas on the horns o' th' moon,
    And with those hands that grasp'd the heaviest club
    Subdue my worthiest self. The witch shall die.
    To the young Roman boy she hath sold me, and I fall
    Under this plot. She dies for't. Eros, ho!              Exit

ACT_4|SC_13
                          SCENE XIII.
               Alexandria. CLEOPATRA's palace

      Enter CLEOPATRA, CHARMIAN, IRAS, and MARDIAN

  CLEOPATRA. Help me, my women. O, he is more mad
    Than Telamon for his shield; the boar of Thessaly
    Was never so emboss'd.
  CHARMIAN. To th'monument!
    There lock yourself, and send him word you are dead.
    The soul and body rive not more in parting
    Than greatness going off.
  CLEOPATRA. To th' monument!
    Mardian, go tell him I have slain myself;
    Say that the last I spoke was 'Antony'
    And word it, prithee, piteously. Hence, Mardian,
    And bring me how he takes my death. To th' monument!
                                                          Exeunt

ACT_4|SC_14
                          SCENE XIV.
                     CLEOPATRA'S palace

                   Enter ANTONY and EROS

  ANTONY. Eros, thou yet behold'st me?
  EROS. Ay, noble lord.
  ANTONY. Sometime we see a cloud that's dragonish;
    A vapour sometime like a bear or lion,
    A tower'd citadel, a pendent rock,
    A forked mountain, or blue promontory
    With trees upon't that nod unto the world
    And mock our eyes with air. Thou hast seen these signs;
    They are black vesper's pageants.
  EROS. Ay, my lord.
  ANTONY. That which is now a horse, even with a thought
    The rack dislimns, and makes it indistinct,
    As water is in water.
  EROS. It does, my lord.
  ANTONY. My good knave Eros, now thy captain is
    Even such a body. Here I am Antony;
    Yet cannot hold this visible shape, my knave.  
    I made these wars for Egypt; and the Queen-
    Whose heart I thought I had, for she had mine,
    Which, whilst it was mine, had annex'd unto't
    A million moe, now lost- she, Eros, has
    Pack'd cards with Caesar, and false-play'd my glory
    Unto an enemy's triumph.
    Nay, weep not, gentle Eros; there is left us
    Ourselves to end ourselves.

                         Enter MARDIAN

    O, thy vile lady!
    She has robb'd me of my sword.
  MARDIAN. No, Antony;
    My mistress lov'd thee, and her fortunes mingled
    With thine entirely.
  ANTONY. Hence, saucy eunuch; peace!
    She hath betray'd me, and shall die the death.
  MARDIAN. Death of one person can be paid but once,
    And that she has discharg'd. What thou wouldst do  
    Is done unto thy hand. The last she spake
    Was 'Antony! most noble Antony!'
    Then in the midst a tearing groan did break
    The name of Antony; it was divided
    Between her heart and lips. She rend'red life,
    Thy name so buried in her.
  ANTONY. Dead then?
  MARDIAN. Dead.
  ANTONY. Unarm, Eros; the long day's task is done,
    And we must sleep. That thou depart'st hence safe
    Does pay thy labour richly. Go.                 Exit MARDIAN
    Off, pluck off!
    The sevenfold shield of Ajax cannot keep
    The battery from my heart. O, cleave, my sides!
    Heart, once be stronger than thy continent,
    Crack thy frail case. Apace, Eros, apace.-
    No more a soldier. Bruised pieces, go;
    You have been nobly borne.- From me awhile.        Exit EROS
    I will o'ertake thee, Cleopatra, and
    Weep for my pardon. So it must be, for now  
    All length is torture. Since the torch is out,
    Lie down, and stray no farther. Now all labour
    Mars what it does; yea, very force entangles
    Itself with strength. Seal then, and all is done.
    Eros!- I come, my queen.- Eros!- Stay for me;
    Where souls do couch on flowers, we'll hand in hand,
    And with our sprightly port make the ghosts gaze.
    Dido and her Aeneas shall want troops,
    And all the haunt be ours.- Come, Eros, Eros!

                       Re-enter EROS

  EROS. What would my lord?
  ANTONY. Since Cleopatra died,
    I have liv'd in such dishonour that the gods
    Detest my baseness. I, that with my sword
    Quarter'd the world, and o'er green Neptune's back
    With ships made cities, condemn myself to lack
    The courage of a woman; less noble mind
    Than she which by her death our Caesar tells  
    'I am conqueror of myself.' Thou art sworn, Eros,
    That, when the exigent should come- which now
    Is come indeed- when I should see behind me
    Th' inevitable prosecution of
    Disgrace and horror, that, on my command,
    Thou then wouldst kill me. Do't; the time is come.
    Thou strik'st not me; 'tis Caesar thou defeat'st.
    Put colour in thy cheek.
  EROS. The gods withhold me!
    Shall I do that which all the Parthian darts,
    Though enemy, lost aim and could not?
  ANTONY. Eros,
    Wouldst thou be window'd in great Rome and see
    Thy master thus with pleach'd arms, bending down
    His corrigible neck, his face subdu'd
    To penetrative shame, whilst the wheel'd seat
    Of fortunate Caesar, drawn before him, branded
    His baseness that ensued?
  EROS. I would not see't.
  ANTONY. Come, then; for with a wound I must be cur'd.  
    Draw that thy honest sword, which thou hast worn
    Most useful for thy country.
  EROS. O, sir, pardon me!
  ANTONY. When I did make thee free, swor'st thou not then
    To do this when I bade thee? Do it at once,
    Or thy precedent services are all
    But accidents unpurpos'd. Draw, and come.
  EROS. Turn from me then that noble countenance,
    Wherein the worship of the whole world lies.
  ANTONY. Lo thee!                            [Turning from him]
  EROS. My sword is drawn.
  ANTONY. Then let it do at once
    The thing why thou hast drawn it.
  EROS. My dear master,
    My captain and my emperor, let me say,
    Before I strike this bloody stroke, farewell.
  ANTONY. 'Tis said, man; and farewell.
  EROS. Farewell, great chief. Shall I strike now?
  ANTONY. Now, Eros.
  EROS. Why, there then! Thus do I escape the sorrow  
    Of Antony's death.                            [Kills himself
  ANTONY. Thrice nobler than myself!
    Thou teachest me, O valiant Eros, what
    I should, and thou couldst not. My queen and Eros
    Have, by their brave instruction, got upon me
    A nobleness in record. But I will be
    A bridegroom in my death, and run into't
    As to a lover's bed. Come, then; and, Eros,
    Thy master dies thy scholar. To do thus
                                          [Falling on his sword]
    I learn'd of thee. How? not dead? not dead?-
    The guard, ho! O, dispatch me!

                  Enter DERCETAS and a guard

  FIRST GUARD. What's the noise?
  ANTONY. I have done my work ill, friends. O, make an end
    Of what I have begun.
  SECOND GUARD. The star is fall'n.
  FIRST GUARD. And time is at his period.  
  ALL. Alas, and woe!
  ANTONY. Let him that loves me, strike me dead.
  FIRST GUARD. Not I.
  SECOND GUARD. Nor I.
  THIRD GUARD. Nor any one.                         Exeunt guard
  DERCETAS. Thy death and fortunes bid thy followers fly.
    This sword but shown to Caesar, with this tidings,
    Shall enter me with him.

                       Enter DIOMEDES

  DIOMEDES. Where's Antony?
  DERCETAS. There, Diomed, there.
  DIOMEDES. Lives he?
    Wilt thou not answer, man?                     Exit DERCETAS
  ANTONY. Art thou there, Diomed? Draw thy sword and give me
    Sufficing strokes for death.
  DIOMEDES. Most absolute lord,
    My mistress Cleopatra sent me to thee.
  ANTONY. When did she send thee?  
  DIOMEDES. Now, my lord.
  ANTONY. Where is she?
  DIOMEDES. Lock'd in her monument. She had a prophesying fear
    Of what hath come to pass; for when she saw-
    Which never shall be found- you did suspect
    She had dispos'd with Caesar, and that your rage
    Would not be purg'd, she sent you word she was dead;
    But fearing since how it might work, hath sent
    Me to proclaim the truth; and I am come,
    I dread, too late.
  ANTONY. Too late, good Diomed. Call my guard, I prithee.
  DIOMEDES. What, ho! the Emperor's guard! The guard, what ho!
    Come, your lord calls!

             Enter four or five of the guard of ANTONY

  ANTONY. Bear me, good friends, where Cleopatra bides;
    'Tis the last service that I shall command you.
  FIRST GUARD. Woe, woe are we, sir, you may not live to wear
    All your true followers out.  
  ALL. Most heavy day!
  ANTONY. Nay, good my fellows, do not please sharp fate
    To grace it with your sorrows. Bid that welcome
    Which comes to punish us, and we punish it,
    Seeming to bear it lightly. Take me up.
    I have led you oft; carry me now, good friends,
    And have my thanks for all.           Exeunt, hearing ANTONY
ACT_4|SC_15
                         SCENE XV.
                   Alexandria. A monument

      Enter CLEOPATRA and her maids aloft, with CHARMIAN
                         and IRAS

  CLEOPATRA. O Charmian, I will never go from hence!
  CHARMIAN. Be comforted, dear madam.
  CLEOPATRA. No, I will not.
    All strange and terrible events are welcome,
    But comforts we despise; our size of sorrow,
    Proportion'd to our cause, must be as great
    As that which makes it.

                   Enter DIOMEDES, below

    How now! Is he dead?
  DIOMEDES. His death's upon him, but not dead.
    Look out o' th' other side your monument;
    His guard have brought him thither.

            Enter, below, ANTONY, borne by the guard  

  CLEOPATRA. O sun,
    Burn the great sphere thou mov'st in! Darkling stand
    The varying shore o' th' world. O Antony,
    Antony, Antony! Help, Charmian; help, Iras, help;
    Help, friends below! Let's draw him hither.
  ANTONY. Peace!
    Not Caesar's valour hath o'erthrown Antony,
    But Antony's hath triumph'd on itself.
  CLEOPATRA. So it should be, that none but Antony
    Should conquer Antony; but woe 'tis so!
  ANTONY. I am dying, Egypt, dying; only
    I here importune death awhile, until
    Of many thousand kisses the poor last
    I lay upon thy lips.
  CLEOPATRA. I dare not, dear.
    Dear my lord, pardon! I dare not,
    Lest I be taken. Not th' imperious show
    Of the full-fortun'd Caesar ever shall
    Be brooch'd with me. If knife, drugs, serpents, have  
    Edge, sting, or operation, I am safe.
    Your wife Octavia, with her modest eyes
    And still conclusion, shall acquire no honour
    Demuring upon me. But come, come, Antony-
    Help me, my women- we must draw thee up;
    Assist, good friends.
  ANTONY. O, quick, or I am gone.
  CLEOPATRA. Here's sport indeed! How heavy weighs my lord!
    Our strength is all gone into heaviness;
    That makes the weight. Had I great Juno's power,
    The strong-wing'd Mercury should fetch thee up,
    And set thee by Jove's side. Yet come a little.
    Wishers were ever fools. O come, come,
                          [They heave ANTONY aloft to CLEOPATRA]
    And welcome, welcome! Die where thou hast liv'd.
    Quicken with kissing. Had my lips that power,
    Thus would I wear them out.
  ALL. A heavy sight!
  ANTONY. I am dying, Egypt, dying.
    Give me some wine, and let me speak a little.  
  CLEOPATRA. No, let me speak; and let me rail so high
    That the false huswife Fortune break her wheel,
    Provok'd by my offence.
  ANTONY. One word, sweet queen:
    Of Caesar seek your honour, with your safety. O!
  CLEOPATRA. They do not go together.
  ANTONY. Gentle, hear me:
    None about Caesar trust but Proculeius.
  CLEOPATRA. My resolution and my hands I'll trust;
    None about Caesar
  ANTONY. The miserable change now at my end
    Lament nor sorrow at; but please your thoughts
    In feeding them with those my former fortunes
    Wherein I liv'd the greatest prince o' th' world,
    The noblest; and do now not basely die,
    Not cowardly put off my helmet to
    My countryman- a Roman by a Roman
    Valiantly vanquish'd. Now my spirit is going
    I can no more.
  CLEOPATRA. Noblest of men, woo't die?  
    Hast thou no care of me? Shall I abide
    In this dull world, which in thy ab'''

print(len_safe_get_embedding(text))